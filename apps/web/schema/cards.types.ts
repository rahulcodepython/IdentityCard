import { z } from "zod"

// Mirrors apps/server/internal/modules/cards/dto.go.
export const resendResponseSchema = z.object({ message: z.string() })
