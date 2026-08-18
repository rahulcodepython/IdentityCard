"use server"

import { redirect } from "next/navigation"

import { register } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { type RegisterInput, registerSchema } from "@/lib/validation/auth"

export type RegisterActionResult = { error: string } | undefined

export async function registerAction(
  input: RegisterInput
): Promise<RegisterActionResult> {
  const parsed = registerSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Please check the highlighted fields." }
  }

  try {
    await register(parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect("/dashboard")
}
