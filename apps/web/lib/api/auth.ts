import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  meResponseSchema,
  messageResponseSchema,
  type RegisterInput,
  registerResponseSchema,
  registerSchema,
  type SendOtpInput,
  sendOtpSchema,
  type VerifyOtpInput,
  verifyOtpSchema,
  type VerifyTotpInput,
  verifyTotpSchema,
} from "@/lib/validation/auth"

export async function register(input: RegisterInput) {
  const body = registerSchema.parse(input)
  const data = await apiFetch("/auth/register", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return registerResponseSchema.parse(data)
}

export async function sendOtp(input: SendOtpInput) {
  const body = sendOtpSchema.parse(input)
  const data = await apiFetch("/auth/otp/send", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return messageResponseSchema.parse(data)
}

export async function verifyOtp(input: VerifyOtpInput) {
  const body = verifyOtpSchema.parse(input)
  const data = await apiFetch("/auth/otp/verify", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return messageResponseSchema.parse(data)
}

export async function verifyTotp(input: VerifyTotpInput) {
  const body = verifyTotpSchema.parse(input)
  const data = await apiFetch("/auth/totp/verify", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return messageResponseSchema.parse(data)
}

export async function logout() {
  const data = await apiFetch("/auth/logout", { method: "POST" })
  return messageResponseSchema.parse(data)
}

export async function me() {
  const data = await apiFetch("/auth/me")
  return meResponseSchema.parse(data)
}
