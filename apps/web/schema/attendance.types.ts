import { z } from "zod"

// Mirrors apps/server/internal/modules/attendance/dto.go's RosterEntry —
// one (person, date) pair they were permitted to enter on, present or not.

export const attendanceStatusSchema = z.enum(["early", "on_time", "late"])

export const rosterEntrySchema = z.object({
  person_id: z.string().uuid(),
  person_name: z.string(),
  person_email: z.string(),
  date: z.string(),
  attended: z.boolean(),
  entry_at: z.string().nullable(),
  entry_status: attendanceStatusSchema.nullable(),
  exit_at: z.string().nullable(),
  exit_status: attendanceStatusSchema.nullable(),
})
export type RosterEntry = z.infer<typeof rosterEntrySchema>

export const rosterResponseSchema = z
  .array(rosterEntrySchema)
  .nullable()
  .transform((v) => v ?? [])
