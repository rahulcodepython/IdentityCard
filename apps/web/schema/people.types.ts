import { z } from "zod";

export const createPersonSchema = z.object({
    email: z
        .string()
        .min(1, "Email is required")
        .email("Enter a valid email address"),
    mobile: z.string().min(3, "Enter a mobile number").max(32),
    name: z.string().min(1, "Enter a name").max(200),
    image_url: z.string().max(2048).default(""),
    age: z.number().int().min(0).max(150).optional(),
    gender: z.string().max(50).default(""),
    sub_event_ids: z.array(z.string().uuid()).default([]),
});
export type CreatePersonInput = z.infer<typeof createPersonSchema>;

export const updatePersonSchema = createPersonSchema.omit({
    email: true,
    mobile: true,
});
export type UpdatePersonInput = z.infer<typeof updatePersonSchema>;

export const personResponseSchema = z.object({
    id: z.string().uuid(),
    email: z.string(),
    mobile: z.string(),
    name: z.string(),
    image_url: z.string().nullable().optional(),
    age: z.number().nullable().optional(),
    gender: z.string().nullable().optional(),
    joined_at: z.string().nullable().optional(),
    card_sent_at: z.string().nullable().optional(),
    sub_event_ids: z.array(z.string().uuid()),
});
export type Person = z.infer<typeof personResponseSchema>;

export const peopleListResponseSchema = z.array(personResponseSchema);

export const importSummarySchema = z.object({
    inserted: z.number(),
    updated: z.number(),
    skipped: z.number(),
    errors: z.array(z.object({ row: z.number(), message: z.string() })),
});
export type ImportSummary = z.infer<typeof importSummarySchema>;
