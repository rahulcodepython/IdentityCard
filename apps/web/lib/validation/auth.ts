import { z } from "zod"

// Mirrors apps/server/internal/modules/auth/dto.go — update both sides
// together when a field changes.

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, "Email is required")
    .email("Enter a valid email address"),
  password: z.string().min(1, "Password is required"),
})
export type LoginInput = z.infer<typeof loginSchema>

export const registerSchema = z.object({
  name: z.string().min(2, "Enter your name").max(120),
  email: z
    .string()
    .min(1, "Email is required")
    .email("Enter a valid email address"),
  password: z.string().min(8, "Password must be at least 8 characters").max(72),
  organization_name: z
    .string()
    .min(2, "Enter your organization's name")
    .max(120),
  plan_code: z.string().min(1, "Choose a plan"),
})
export type RegisterInput = z.infer<typeof registerSchema>

export const messageResponseSchema = z.object({
  message: z.string(),
})

export const roleSchema = z.enum(["super_admin", "admin", "scanner"])
export type Role = z.infer<typeof roleSchema>

export const meResponseSchema = z.object({
  user_id: z.string().uuid(),
  email: z.string().email(),
  name: z.string(),
  organization_id: z.string().uuid(),
  organization_name: z.string(),
  roles: z.array(roleSchema),
})
export type MeResponse = z.infer<typeof meResponseSchema>
