import { z } from "zod"

// Mirrors apps/server/internal/modules/devices/dto.go (admin side — the
// client-side pairing schemas live in lib/validation/scanner.ts).

export const createDeviceSchema = z.object({
  name: z.string().min(1, "Enter a device name").max(100),
})
export type CreateDeviceInput = z.infer<typeof createDeviceSchema>

export const deviceStatusSchema = z.enum(["pending", "verified", "revoked"])

export const deviceResponseSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  status: deviceStatusSchema,
  created_at: z.string(),
  verified_at: z.string().nullable(),
})
export type Device = z.infer<typeof deviceResponseSchema>

export const devicesListResponseSchema = z.array(deviceResponseSchema)

export const createDeviceResponseSchema = deviceResponseSchema.extend({
  otp_code: z.string(),
  otp_expires_at: z.string(),
})
export type CreateDeviceResult = z.infer<typeof createDeviceResponseSchema>
