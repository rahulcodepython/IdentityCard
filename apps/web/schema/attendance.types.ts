import { z } from "zod";

export const attendanceStatusSchema = z.enum(["early", "on_time", "late"]);

export const rosterEntrySchema = z.object({
    person_id: z.string().uuid(),
    person_name: z.string(),
    person_email: z.string(),
    date: z.string(),
    attended: z.boolean(),
    entry_at: z.string().nullable().optional(),
    entry_status: attendanceStatusSchema.nullable().optional(),
    exit_at: z.string().nullable().optional(),
    exit_status: attendanceStatusSchema.nullable().optional(),
});
export type RosterEntry = z.infer<typeof rosterEntrySchema>;

export const rosterResponseSchema = z
    .array(rosterEntrySchema)
    .nullable()
    .transform((v) => v ?? []);
