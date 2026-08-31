"use client";

import { apiRequest } from "@/react-query/client";
import { useAppQuery } from "@/react-query/query";
import { type RosterFilter, queryKeys } from "@/react-query/query-keys";
import {
    type RosterEntry,
    rosterResponseSchema,
} from "@/schema/attendance.types";

export function useAttendanceRosterQuery(eventId: string, filter: RosterFilter = {}, enabled = true) {
    const params = new URLSearchParams();
    if (filter.subEventId) params.set("sub_event_id", filter.subEventId);
    if (filter.date) params.set("date", filter.date);
    if (filter.status) params.set("status", filter.status);
    if (filter.attended !== undefined) params.set("attended", String(filter.attended));
    const qs = params.toString();

    return useAppQuery<RosterEntry[]>(
        queryKeys.roster(eventId, filter),
        () =>
            apiRequest(
                { url: `/events/${eventId}/attendance${qs ? `?${qs}` : ""}`, method: "GET" },
                rosterResponseSchema
            ),
        { enabled: enabled && !!eventId }
    );
}

