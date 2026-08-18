import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  type CreateDeviceInput,
  createDeviceResponseSchema,
  createDeviceSchema,
  devicesListResponseSchema,
} from "@/lib/validation/devices"

export async function listDevices() {
  const data = await apiFetch("/devices")
  return devicesListResponseSchema.parse(data)
}

export async function createDevice(input: CreateDeviceInput) {
  const body = createDeviceSchema.parse(input)
  const data = await apiFetch("/devices", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return createDeviceResponseSchema.parse(data)
}

export async function revokeDevice(id: string) {
  await apiFetch(`/devices/${id}/revoke`, { method: "POST" })
}
