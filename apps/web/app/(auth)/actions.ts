"use server"

import { redirect } from "next/navigation"

import { sendOtp, verifyOtp, verifyTotp } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import {
  type SendOtpInput,
  sendOtpSchema,
  type VerifyOtpInput,
  verifyOtpSchema,
  type VerifyTotpInput,
  verifyTotpSchema,
} from "@/lib/validation/auth"

export type VerifyActionResult = { error: string } | undefined

// Shared by the login and register flows — both end at the same
// email-code/authenticator verification screen (see components/verification.tsx).
export async function sendOtpAction(
  input: SendOtpInput
): Promise<VerifyActionResult> {
  const parsed = sendOtpSchema.safeParse(input)
  if (!parsed.success) return { error: "Enter a valid email address." }

  try {
    await sendOtp(parsed.data)
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message }
    return { error: "Something went wrong. Please try again." }
  }
  return undefined
}

export async function verifyOtpAction(
  input: VerifyOtpInput
): Promise<VerifyActionResult> {
  const parsed = verifyOtpSchema.safeParse(input)
  if (!parsed.success) return { error: "Enter the 6-digit code." }

  try {
    await verifyOtp(parsed.data)
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message }
    return { error: "Something went wrong. Please try again." }
  }
  redirect("/dashboard")
}

export async function verifyTotpAction(
  input: VerifyTotpInput
): Promise<VerifyActionResult> {
  const parsed = verifyTotpSchema.safeParse(input)
  if (!parsed.success) return { error: "Enter the 6-digit code." }

  try {
    await verifyTotp(parsed.data)
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message }
    return { error: "Something went wrong. Please try again." }
  }
  redirect("/dashboard")
}
