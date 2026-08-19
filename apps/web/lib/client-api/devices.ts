import { apiRequest } from "@/react-query/client"
import {
  type CreateDeviceInput,
  createDeviceResponseSchema,
  createDeviceSchema,
  devicesListResponseSchema,
} from "@/schema/devices.types"
import { z } from "zod"

export async function listDevices() {
  return apiRequest({ url: "/devices", method: "GET" }, devicesListResponseSchema)
}

export async function createDevice(input: CreateDeviceInput) {
  const body = createDeviceSchema.parse(input)
  return apiRequest(
    { url: "/devices", method: "POST", data: body },
    createDeviceResponseSchema
  )
}

export async function revokeDevice(id: string) {
  await apiRequest({ url: `/devices/${id}/revoke`, method: "POST" }, z.unknown())
}

export async function removeDevice(id: string) {
  await apiRequest({ url: `/devices/${id}`, method: "DELETE" }, z.unknown())
}