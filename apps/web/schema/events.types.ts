import { z } from "zod";

export const eventDayInputSchema = z.object({
    date: z.string().min(1, "Date is required"),
    entry_time: z.string().min(1, "Entry time is required"),
    exit_time: z.string().min(1, "Exit time is required"),
});
export type EventDayInput = z.infer<typeof eventDayInputSchema>;

export const eventDayResponseSchema = z.object({
    date: z.string(),
    entry_time: z.string(),
    exit_time: z.string(),
});
export type EventDay = z.infer<typeof eventDayResponseSchema>;

export const eventTypeSchema = z.enum(["flash", "standard", "grouped"]);
export type EventType = z.infer<typeof eventTypeSchema>;

export const eventStatusSchema = z.enum(["draft", "published"]);
export type EventStatus = z.infer<typeof eventStatusSchema>;

export const createEventSchema = z.object({
    event_type: eventTypeSchema,
    name: z.string().min(2, "Enter an event name").max(200),
    venue: z.string().max(300).default(""),
    organizer_name: z.string().max(200).default(""),
    days: z.array(eventDayInputSchema).default([]),
    range_start: z.string().default(""),
    range_end: z.string().default(""),
});
export type CreateEventInput = z.infer<typeof createEventSchema>;

export const updateEventSchema = z.object({
    name: z.string().min(2, "Enter an event name").max(200),
    venue: z.string().max(300).default(""),
    organizer_name: z.string().max(200).default(""),
    days: z.array(eventDayInputSchema).default([]),
    range_start: z.string().default(""),
    range_end: z.string().default(""),
});
export type UpdateEventInput = z.infer<typeof updateEventSchema>;

export const eventSummarySchema = z.object({
    id: z.string().uuid(),
    event_type: eventTypeSchema,
    status: eventStatusSchema,
    start_date: z.string(),
    end_date: z.string(),
    name: z.string(),
    venue: z.string().nullable().optional(),
    organizer_name: z.string().nullable().optional(),
    has_image: z.boolean(),
    has_organizer_signature: z.boolean(),
    published_at: z.string().nullable().optional(),
});
export type EventSummary = z.infer<typeof eventSummarySchema>;

export const eventResponseSchema = eventSummarySchema.extend({
    days: z.array(eventDayResponseSchema),
});
export type EventDetail = z.infer<typeof eventResponseSchema>;

export const eventsListResponseSchema = z.array(eventSummarySchema);

export const dayImportRowErrorSchema = z.object({
    row: z.number(),
    message: z.string(),
});

export const dayImportSummarySchema = z.object({
    imported: z.number(),
    skipped: z.number(),
    errors: z.array(dayImportRowErrorSchema),
});
export type DayImportSummary = z.infer<typeof dayImportSummarySchema>;
