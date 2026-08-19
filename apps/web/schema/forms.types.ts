import { z } from "zod"

// Mirrors apps/server/internal/modules/forms/dto.go.

export const createFormSchema = z.object({
  sub_event_id: z.string().uuid().optional(),
  capacity: z.number().int().min(1).optional(),
})
export type CreateFormInput = z.infer<typeof createFormSchema>

export const updateFormSchema = z.object({
  capacity: z.number().int().min(1).optional(),
  is_active: z.boolean(),
})
export type UpdateFormInput = z.infer<typeof updateFormSchema>

export const formResponseSchema = z.object({
  id: z.string().uuid(),
  token: z.string(),
  sub_event_id: z.string().uuid().nullable(),
  capacity: z.number().nullable(),
  submissions_count: z.number(),
  is_active: z.boolean(),
})
export type EventForm = z.infer<typeof formResponseSchema>

export const formsListResponseSchema = z.array(formResponseSchema)

// Public — served with no authentication, token possession is the access
// control (see apps/server/internal/modules/forms/handler.go GetPublic/Submit).

export const publicFormResponseSchema = z.object({
  event_name: z.string(),
  sub_event_name: z.string().nullable(),
  is_open: z.boolean(),
})
export type PublicForm = z.infer<typeof publicFormResponseSchema>

export const submitFormSchema = z.object({
  email: z
    .string()
    .min(1, "Email is required")
    .email("Enter a valid email address"),
  mobile: z.string().min(3, "Enter a mobile number").max(32),
  name: z.string().min(1, "Enter your name").max(200),
  image_url: z.string().max(2048).default(""),
  age: z.number().int().min(0).max(150).optional(),
  gender: z.string().max(50).default(""),
})
export type SubmitFormInput = z.infer<typeof submitFormSchema>

export const messageResponseSchema = z.object({ message: z.string() })
