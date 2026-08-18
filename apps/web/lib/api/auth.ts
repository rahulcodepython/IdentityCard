import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  type LoginInput,
  loginSchema,
  meResponseSchema,
  messageResponseSchema,
  type RegisterInput,
  registerSchema,
} from "@/lib/validation/auth"

export async function login(input: LoginInput) {
  const body = loginSchema.parse(input)
  const data = await apiFetch("/auth/login", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return messageResponseSchema.parse(data)
}

export async function register(input: RegisterInput) {
  const body = registerSchema.parse(input)
  const data = await apiFetch("/auth/register", {
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
