import { z } from "zod";

export const EventFormStatusSchema = z.enum(["waiting", "live"]);
export type EventFormStatus = z.infer<typeof EventFormStatusSchema>;

export const EventFormSchema = z.object({
    id: z.string(),
    event_id: z.string(),
    form_id: z.string(),
    max_applicants: z.number().default(-1),
    expires_at: z.string(),
    status: EventFormStatusSchema.default("waiting"),
    created_at: z.string().optional(),
    updated_at: z.string().optional(),
});

export type EventForm = z.infer<typeof EventFormSchema>;

export const AssignedFormSummarySchema = z.object({
    id: z.string(),
    name: z.string(),
    fields_count: z.number().default(0),
    is_published: z.boolean().default(false),
});

export type AssignedFormSummary = z.infer<typeof AssignedFormSummarySchema>;

export const EventSharingResponseSchema = z.object({
    event_form: EventFormSchema.nullish(),
    assigned_form: AssignedFormSummarySchema.nullish(),
    total_applicants: z.number().nullish().default(0),
    can_change_form: z.boolean().nullish().default(true),
});

export type EventSharingResponse = z.infer<typeof EventSharingResponseSchema>;


export const UpdateEventSharingSchema = z.object({
    form_id: z.string().min(1, "Form is required"),
    max_applicants: z.coerce.number().int().min(-1, "Must be -1 or greater"),
    expires_at: z.string().min(1, "Expiry date is required"),
    status: EventFormStatusSchema.default("waiting"),
});

export type UpdateEventSharingInput = z.infer<typeof UpdateEventSharingSchema>;

export const SharingConfigFormSchema = z.object({
    form_id: z.string().min(1, "Please select a published form"),
    max_applicants: z.coerce
        .number({ invalid_type_error: "Must be a valid number" })
        .int("Must be an integer")
        .refine((val) => val === -1 || val >= 1, {
            message: "Max applicants must be -1 (unlimited) or at least 1",
        }),
    expires_at: z
        .string()
        .min(1, "Please set an expiration date and time")
        .refine((val) => !isNaN(new Date(val).getTime()), {
            message: "Invalid expiration date",
        })
        .refine((val) => new Date(val).getTime() > Date.now(), {
            message: "Expiry date must be in the future",
        }),
    status: EventFormStatusSchema.default("waiting"),
});

export type SharingConfigFormValues = z.infer<typeof SharingConfigFormSchema>;


