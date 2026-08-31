import { z } from "zod";

export const createSubEventSchema = z.object({
    name: z.string().min(2, "Enter a sub-event name").max(200),
    date: z.string().min(1, "Date is required"),
    entry_time: z.string().min(1, "Entry time is required"),
    exit_time: z.string().min(1, "Exit time is required"),
});
export type CreateSubEventInput = z.infer<typeof createSubEventSchema>;

export const updateSubEventSchema = z.object({
    name: z.string().min(2, "Enter a sub-event name").max(200),
    date: z.string().min(1, "Date is required"),
    entry_time: z.string().min(1, "Entry time is required"),
    exit_time: z.string().min(1, "Exit time is required"),
});
export type UpdateSubEventInput = z.infer<typeof updateSubEventSchema>;

export const subEventResponseSchema = z.object({
    id: z.string().uuid(),
    name: z.string(),
    date: z.string(),
    entry_time: z.string(),
    exit_time: z.string(),
});
export type SubEvent = z.infer<typeof subEventResponseSchema>;

export const subEventsListResponseSchema = z.array(subEventResponseSchema);
