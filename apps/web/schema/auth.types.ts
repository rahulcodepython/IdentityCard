import { z } from "zod"

export const registerSchema = z.object({
    name: z.string().min(2, "Enter your name").max(120),
    organizationName: z
        .string()
        .min(2, "Enter your organization name")
        .max(120),
    email: z
        .string()
        .min(1, "Email is required")
        .email("Enter a valid email address"),
})
export type RegisterInput = z.infer<typeof registerSchema>

export const sendOtpSchema = z.object({
    email: z
        .string()
        .min(1, "Email is required")
        .email("Enter a valid email address"),
})
export type SendOtpInput = z.infer<typeof sendOtpSchema>

export const totpLoginSchema = z.object({
    email: z
        .string()
        .min(1, "Email is required")
        .email("Enter a valid email address"),
    code: z.string().length(6, "Authenticator code must be exactly 6 digits"),
})
export type TotpLoginInput = z.infer<typeof totpLoginSchema>

export const messageResponseSchema = z.object({
    message: z.string(),
})

export const userRoleSchema = z.enum(["admin", "user"])
export type UserRole = z.infer<typeof userRoleSchema>

export const roleSchema = z.enum(["admin", "member"])
export type Role = z.infer<typeof roleSchema>

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
