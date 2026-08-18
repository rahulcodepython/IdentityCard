"use server"

import { redirect } from "next/navigation"

import { resendCard } from "@/lib/api/cards"
import { ApiError } from "@/lib/api/client"
import { deletePerson, updatePerson } from "@/lib/api/people"
import {
  type UpdatePersonInput,
  updatePersonSchema,
} from "@/lib/validation/people"

export type PersonActionResult = { error: string } | undefined

export async function updatePersonAction(
  eventId: string,
  personId: string,
  input: UpdatePersonInput
): Promise<PersonActionResult> {
  const parsed = updatePersonSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }

  try {
    await updatePerson(eventId, personId, parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect(`/dashboard/events/${eventId}/people`)
}

export async function deletePersonAction(
  eventId: string,
  personId: string
): Promise<PersonActionResult> {
  try {
    await deletePerson(eventId, personId)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect(`/dashboard/events/${eventId}/people`)
}

export async function resendCardAction(
  eventId: string,
  personId: string
): Promise<PersonActionResult> {
  try {
    await resendCard(eventId, personId)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
}
