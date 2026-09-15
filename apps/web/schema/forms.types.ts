import { z } from "zod";

import { PaginatedResponseZod, type PaginatedResponse } from "@/schema/common.types";

// =====================================================================
// Supported Field Types
// =====================================================================

export const FORM_FIELD_TYPES = [
    "text",
    "email",
    "number",
    "url",
    "checkbox",
    "radio",
    "switch",
    "date",
    "time",
    "month",
    "week",
    "file",
] as const;

export type FormFieldType = (typeof FORM_FIELD_TYPES)[number];

// =====================================================================
// Field Option & Validation Schemas
// =====================================================================

export const FieldOptionSchema = z.object({
    id: z.string(),
    label: z.string(),
    value: z.string(),
});

export type FieldOption = z.infer<typeof FieldOptionSchema>;

export const FieldValidationSchema = z.object({
    min: z.number().optional().nullable(),
    max: z.number().optional().nullable(),
    pattern: z.string().optional().nullable(),
    accept: z.string().optional().nullable(),
    max_file_size_mb: z.number().optional().nullable(),
    multiple: z.boolean().optional().nullable(),
});

export type FieldValidation = z.infer<typeof FieldValidationSchema>;

// =====================================================================
// Form Field Schema (Stored in Postgres JSONB)
// =====================================================================

export const FormFieldSchema = z.object({
    id: z.string(),
    key: z.string(),
    label: z.string(),
    type: z.enum(FORM_FIELD_TYPES),
    required: z.boolean().default(false),
    placeholder: z.string().default(""),
    is_system: z.boolean().default(false),
    options: z.array(FieldOptionSchema).default([]),
    validation: FieldValidationSchema.optional().nullable(),
});

export type FormField = z.infer<typeof FormFieldSchema>;

// =====================================================================
// Form Entity Schemas
// =====================================================================

export const FormSchema = z.object({
    id: z.string(),
    name: z.string(),
    fields: z.array(FormFieldSchema),
    created_at: z.string(),
    updated_at: z.string(),
});

export type Form = z.infer<typeof FormSchema>;

export const PaginatedFormsSchema = PaginatedResponseZod(FormSchema);
export type PaginatedForms = PaginatedResponse<Form>;

// =====================================================================
// Request Payload Schemas
// =====================================================================

export const CreateFormSchema = z.object({
    name: z.string().min(1, "Form name is required").trim(),
});

export type CreateFormInput = z.infer<typeof CreateFormSchema>;

export const UpdateFormSchema = z.object({
    name: z.string().min(1, "Form name is required").trim(),
});

export type UpdateFormInput = z.infer<typeof UpdateFormSchema>;

export const UpdateFormFieldsSchema = z.object({
    fields: z.array(FormFieldSchema),
});

export type UpdateFormFieldsInput = z.infer<typeof UpdateFormFieldsSchema>;
