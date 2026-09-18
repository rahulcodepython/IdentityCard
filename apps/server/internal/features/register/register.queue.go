package register

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"

    "github.com/redis/go-redis/v9"
)

const (
    StreamEventSubmissions = "stream:event_submissions"
    GroupRegisterWorkers   = "register_workers"
)

// luaReserveAndEnqueue atomically checks duplicate email, checks and decrements slot counter,
// and pushes the submission payload to the Redis Stream.
const luaReserveAndEnqueue = `
local slotsKey = KEYS[1]
local emailsKey = KEYS[2]
local streamKey = KEYS[3]

local email = ARGV[1]
local maxApplicants = tonumber(ARGV[2])
local initialRemainingSlots = tonumber(ARGV[3])
local userID = ARGV[4]
local payloadJSON = ARGV[5]

-- 1. Check duplicate email in event set
local isMember = redis.call('SISMEMBER', emailsKey, email)
if isMember == 1 then
    return 'already_registered'
end

-- 2. Check and decrement slot capacity if limit exists
if maxApplicants ~= -1 then
    local exists = redis.call('EXISTS', slotsKey)
    if exists == 0 then
        redis.call('SET', slotsKey, initialRemainingSlots)
        redis.call('EXPIRE', slotsKey, 86400)
    end
    local currentSlots = tonumber(redis.call('GET', slotsKey))
    if currentSlots <= 0 then
        return 'limit_reached'
    end
    redis.call('DECR', slotsKey)
end

-- 3. Mark email in Redis set
redis.call('SADD', emailsKey, email)
redis.call('EXPIRE', emailsKey, 86400)

-- 4. Push to stream
redis.call('XADD', streamKey, '*', 'user_id', userID, 'payload', payloadJSON)

return 'ok'
`

var reserveAndEnqueueScript = redis.NewScript(luaReserveAndEnqueue)

// EnqueueSubmission executes the atomic reservation script in Redis.
// Returns "ok", "already_registered", "limit_reached", or error.
func (s *App) EnqueueSubmission(ctx context.Context, sub *QueuedSubmission, maxApplicants, currentApplicants int) (string, error) {
    if s.RDB == nil {
        return "", fmt.Errorf("redis client not available")
    }

    payloadBytes, err := json.Marshal(sub)
    if err != nil {
        return "", fmt.Errorf("marshal queued submission: %w", err)
    }

    slotsKey := fmt.Sprintf("event:%s:slots", sub.EventID)
    emailsKey := fmt.Sprintf("event:%s:emails", sub.EventID)

    remainingSlots := maxApplicants - currentApplicants
    if remainingSlots < 0 {
        remainingSlots = 0
    }

    keys := []string{slotsKey, emailsKey, StreamEventSubmissions}
    args := []interface{}{
        sub.Email,
        maxApplicants,
        remainingSlots,
        sub.UserID,
        string(payloadBytes),
    }

    result, err := reserveAndEnqueueScript.Run(ctx, s.RDB, keys, args...).Text()
    if err != nil {
        slog.Error("luaReserveAndEnqueue error", "error", err, "event_id", sub.EventID, "email", sub.Email)
        return "", err
    }

    return result, nil
}
