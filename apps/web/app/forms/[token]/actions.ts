"use server"

import { ApiError } from "@/lib/api/client"
import { submitPublicForm } from "@/lib/api/forms"
import { type SubmitFormInput, submitFormSchema } from "@/lib/validation/forms"

export type SubmitActionResult = { error: string } | { success: true }

export async function submitFormAction(
  token: string,
  input: SubmitFormInput
): Promise<SubmitActionResult> {
  const parsed = submitFormSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }

  try {
    await submitPublicForm(token, parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  return { success: true }
}
