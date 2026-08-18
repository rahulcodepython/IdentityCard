"use server"

import { revalidatePath } from "next/cache"

import { ApiError } from "@/lib/api/client"
import { createForm, deleteForm, updateForm } from "@/lib/api/forms"
import {
  type CreateFormInput,
  createFormSchema,
  type UpdateFormInput,
  updateFormSchema,
} from "@/lib/validation/forms"

export type FormActionResult = { error: string } | undefined

export async function createFormAction(
  eventId: string,
  input: CreateFormInput
): Promise<FormActionResult> {
  const parsed = createFormSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }
  try {
    await createForm(eventId, parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
  revalidatePath(`/dashboard/events/${eventId}/forms`)
}

export async function updateFormAction(
  eventId: string,
  id: string,
  input: UpdateFormInput
): Promise<FormActionResult> {
  const parsed = updateFormSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Invalid update" }
  }
  try {
    await updateForm(eventId, id, parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
  revalidatePath(`/dashboard/events/${eventId}/forms`)
}

export async function deleteFormAction(
  eventId: string,
  id: string
): Promise<FormActionResult> {
  try {
    await deleteForm(eventId, id)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
  revalidatePath(`/dashboard/events/${eventId}/forms`)
}
