"use server"

import { redirect } from "next/navigation"

import { login } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { type LoginInput, loginSchema } from "@/lib/validation/auth"

export type LoginActionResult = { error: string } | undefined

// Client-side validation already happened in LoginForm (react-hook-form +
// zodResolver) — this re-validates because a Server Action is a public
// endpoint in disguise and must never trust the client.
export async function loginAction(
  input: LoginInput
): Promise<LoginActionResult> {
  const parsed = loginSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Enter a valid email and password." }
  }

  try {
    await login(parsed.data)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }

  redirect("/dashboard")
}
