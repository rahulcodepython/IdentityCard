import { z } from "zod"

// Mirrors apps/server/internal/modules/auth/dto.go — update both sides
// together when a field changes.

export const registerSchema = z.object({
  name: z.string().min(2, "Enter your name").max(120),
  email: z
    .string()
    .min(1, "Email is required")
    .email("Enter a valid email address"),
  organization_name: z
    .string()
    .min(2, "Enter your organization's name")
    .max(120),
})
export type RegisterInput = z.infer<typeof registerSchema>

export const registerResponseSchema = z.object({
  email: z.string().email(),
  totp_qr_image: z.string(),
  totp_secret: z.string(),
})
export type RegisterResponse = z.infer<typeof registerResponseSchema>

export const sendOtpSchema = z.object({
  email: z.string().min(1, "Email is required").email("Enter a valid email address"),
})
export type SendOtpInput = z.infer<typeof sendOtpSchema>

const codeSchema = z
  .string()
  .length(6, "Enter the 6-digit code")
  .regex(/^\d{6}$/, "Code must be numeric")

export const verifyOtpSchema = z.object({
  email: z.string().email(),
  code: codeSchema,
})
export type VerifyOtpInput = z.infer<typeof verifyOtpSchema>

export const verifyTotpSchema = z.object({
  email: z.string().email(),
  code: codeSchema,
})
export type VerifyTotpInput = z.infer<typeof verifyTotpSchema>

export const messageResponseSchema = z.object({
  message: z.string(),
})

export const roleSchema = z.enum(["super_admin", "admin", "scanner"])
export type Role = z.infer<typeof roleSchema>

// organization_id/organization_name/roles are only present once
// has_organization is true — a Google signup lands with a valid session
// but no org yet, pending /onboarding.
export const meResponseSchema = z.object({
  user_id: z.string().uuid(),
  email: z.string().email(),
  name: z.string(),
  has_organization: z.boolean(),
  organization_id: z.string().uuid().optional(),
  organization_name: z.string().optional(),
  roles: z.array(roleSchema).optional(),
})
export type MeResponse = z.infer<typeof meResponseSchema>
