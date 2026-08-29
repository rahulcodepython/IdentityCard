import { z } from "zod"

// Auth itself is better-auth's (apps/web/lib/auth.ts) — these schemas
// are just this app's own form/DTO shapes around it, not a mirror of a
// Go DTO anymore.

export const registerSchema = z.object({
  name: z.string().min(2, "Enter your name").max(120),
  email: z
    .string()
    .min(1, "Email is required")
    .email("Enter a valid email address"),
})
export type RegisterInput = z.infer<typeof registerSchema>

export const sendOtpSchema = z.object({
  email: z.string().min(1, "Email is required").email("Enter a valid email address"),
})
export type SendOtpInput = z.infer<typeof sendOtpSchema>

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
