import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  type CreateFormInput,
  createFormSchema,
  formResponseSchema,
  formsListResponseSchema,
  messageResponseSchema,
  publicFormResponseSchema,
  type SubmitFormInput,
  submitFormSchema,
  type UpdateFormInput,
  updateFormSchema,
} from "@/lib/validation/forms"

export async function listForms(eventId: string) {
  const data = await apiFetch(`/events/${eventId}/forms`)
  return formsListResponseSchema.parse(data)
}

export async function createForm(eventId: string, input: CreateFormInput) {
  const body = createFormSchema.parse(input)
  const data = await apiFetch(`/events/${eventId}/forms`, {
    method: "POST",
    body: JSON.stringify(body),
  })
  return formResponseSchema.parse(data)
}

export async function updateForm(
  eventId: string,
  id: string,
  input: UpdateFormInput
) {
  const body = updateFormSchema.parse(input)
  const data = await apiFetch(`/events/${eventId}/forms/${id}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  })
  return formResponseSchema.parse(data)
}

export async function deleteForm(eventId: string, id: string) {
  await apiFetch(`/events/${eventId}/forms/${id}`, { method: "DELETE" })
}

// Public — no auth cookie required; apiFetch forwards whatever's present
// (nothing, for a fresh visitor) harmlessly since these endpoints don't
// require it.

export async function getPublicForm(token: string) {
  const data = await apiFetch(`/public/forms/${token}`)
  return publicFormResponseSchema.parse(data)
}

export async function submitPublicForm(token: string, input: SubmitFormInput) {
  const body = submitFormSchema.parse(input)
  const data = await apiFetch(`/public/forms/${token}/submit`, {
    method: "POST",
    body: JSON.stringify(body),
  })
  return messageResponseSchema.parse(data)
}
