import { z } from "zod"

// Mirrors apps/server/internal/modules/devices/dto.go (pair/me) and
// internal/modules/attendance/dto.go (scan) — these are parsed client-side
// (see lib/device-client.ts), not server-only, since a scanner device
// talks to the Go API directly from the browser.

export const pairResponseSchema = z.object({
  device_id: z.string().uuid(),
  organization_name: z.string(),
  key: z.string(),
})
export type PairResponse = z.infer<typeof pairResponseSchema>

export const scannerMeResponseSchema = z.object({
  device_id: z.string().uuid(),
  device_name: z.string(),
  organization_name: z.string(),
})
export type ScannerMe = z.infer<typeof scannerMeResponseSchema>

const scanPersonSchema = z.object({
  id: z.string().uuid(),
  email: z.string(),
  mobile: z.string(),
  name: z.string(),
  image_url: z.string().nullable(),
  age: z.number().nullable(),
  gender: z.string().nullable(),
})

export const scanResponseSchema = z.object({
  direction: z.enum(["entry", "exit", "already_completed"]),
  status: z.enum(["early", "on_time", "late"]).nullable(),
  date: z.string(),
  event_name: z.string(),
  person: scanPersonSchema,
})
export type ScanResponse = z.infer<typeof scanResponseSchema>
