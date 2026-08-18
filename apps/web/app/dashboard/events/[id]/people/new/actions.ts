"use server"

import { redirect } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { createPerson } from "@/lib/api/people"
import {
  type CreatePersonInput,
  createPersonSchema,
} from "@/lib/validation/people"

export type PersonActionResult = { error: string } | undefined

export async function createPersonAction(
  eventId: string,
  input: CreatePersonInput
): Promise<PersonActionResult> {
  const parsed = createPersonSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }

  try {
    await createPerson(eventId, parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect(`/dashboard/events/${eventId}/people`)
}
