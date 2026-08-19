import { z } from "zod"

import { apiRequest } from "@/react-query/client"
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
} from "@/schema/forms.types"

export async function listForms(eventId: string) {
  return apiRequest(
    { url: `/events/${eventId}/forms`, method: "GET" },
    formsListResponseSchema
  )
}

export async function createForm(eventId: string, input: CreateFormInput) {
  const body = createFormSchema.parse(input)
  return apiRequest(
    { url: `/events/${eventId}/forms`, method: "POST", data: body },
    formResponseSchema
  )
}

export async function updateForm(
  eventId: string,
  id: string,
  input: UpdateFormInput
) {
  const body = updateFormSchema.parse(input)
  return apiRequest(
    { url: `/events/${eventId}/forms/${id}`, method: "PATCH", data: body },
    formResponseSchema
  )
}

export async function deleteForm(eventId: string, id: string) {
  await apiRequest(
    { url: `/events/${eventId}/forms/${id}`, method: "DELETE" },
    z.unknown()
  )
}

// Public — no auth cookie required; the axios client sends whatever
// session cookie is present (nothing, for a fresh visitor) harmlessly
// since these endpoints don't require it.

export async function getPublicForm(token: string) {
  return apiRequest(
    { url: `/public/forms/${token}`, method: "GET" },
    publicFormResponseSchema
  )
}

export async function submitPublicForm(token: string, input: SubmitFormInput) {
  const body = submitFormSchema.parse(input)
  return apiRequest(
    { url: `/public/forms/${token}/submit`, method: "POST", data: body },
    messageResponseSchema
  )
}