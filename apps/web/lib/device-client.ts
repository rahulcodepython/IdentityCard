"use client"

// Client-side (not "server-only") on purpose: a scanner-bot device never
// logs in as an org user, so there's no httpOnly cookie to forward from a
// Server Component. Its key lives in localStorage and is sent as a plain
// header directly from the browser to the Go API — see
// apps/server/internal/modules/devices/middleware.go (X-Device-Key).

import {
  pairResponseSchema,
  scannerMeResponseSchema,
  scanResponseSchema,
} from "@/lib/validation/scanner"

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080"
const DEVICE_KEY_STORAGE_KEY = "identitycard_device_key"

export function getDeviceKey(): string | null {
  if (typeof window === "undefined") return null
  return window.localStorage.getItem(DEVICE_KEY_STORAGE_KEY)
}

export function setDeviceKey(key: string) {
  window.localStorage.setItem(DEVICE_KEY_STORAGE_KEY, key)
}

export function clearDeviceKey() {
  window.localStorage.removeItem(DEVICE_KEY_STORAGE_KEY)
}

export class DeviceApiError extends Error {
  status: number
  code: string
  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

async function deviceFetch<T>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const key = getDeviceKey()
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(key ? { "X-Device-Key": key } : {}),
      ...init.headers,
    },
  })

  const body = await res.json().catch(() => ({}))
  if (!res.ok || body.error) {
    throw new DeviceApiError(
      res.status,
      body.error?.code ?? "unknown_error",
      body.error?.message ?? "Request failed"
    )
  }
  return body.data as T
}

export async function pairDevice(otpCode: string) {
  const data = await deviceFetch("/public/devices/pair", {
    method: "POST",
    body: JSON.stringify({ otp_code: otpCode }),
  })
  return pairResponseSchema.parse(data)
}

export async function getScannerMe() {
  const data = await deviceFetch("/scanner/me")
  return scannerMeResponseSchema.parse(data)
}

export async function scanQr(qrToken: string) {
  const data = await deviceFetch("/scanner/scan", {
    method: "POST",
    body: JSON.stringify({ qr_token: qrToken }),
  })
  return scanResponseSchema.parse(data)
}
