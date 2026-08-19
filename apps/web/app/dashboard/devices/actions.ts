"use server"

import { revalidatePath } from "next/cache"

import { ApiError } from "@/lib/api/client"
import { createDevice, removeDevice, revokeDevice } from "@/lib/api/devices"
import {
  type CreateDeviceInput,
  createDeviceSchema,
} from "@/lib/validation/devices"

export type CreateDeviceActionResult =
  | { error: string }
  | { id: string; deviceName: string; otpCode: string; otpExpiresAt: string }

export async function createDeviceAction(
  input: CreateDeviceInput
): Promise<CreateDeviceActionResult> {
  const parsed = createDeviceSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Enter a device name." }
  }

  try {
    const device = await createDevice(parsed.data)
    revalidatePath("/dashboard/devices")
    return {
      id: device.id,
      deviceName: device.name,
      otpCode: device.otp_code,
      otpExpiresAt: device.otp_expires_at,
    }
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
}

export type RevokeDeviceActionResult = { error: string } | undefined

export async function revokeDeviceAction(
  id: string
): Promise<RevokeDeviceActionResult> {
  try {
    await revokeDevice(id)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
  revalidatePath("/dashboard/devices")
}

export async function removeDeviceAction(id: string) {
  try {
    await removeDevice(id)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
  }
  revalidatePath("/dashboard/devices")
}
