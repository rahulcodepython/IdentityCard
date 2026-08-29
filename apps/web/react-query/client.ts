"use client"

import axios, {
  type AxiosError,
  type AxiosRequestConfig,
} from "axios"
import type { ZodType, ZodTypeDef } from "zod"

import { authClient } from "@/lib/auth-client"
import { decodeJwtPayload } from "@/lib/jwt"
import { useSessionStore } from "@/store/session.store"

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080"

type ApiEnvelope<T> = {
  data?: T
  error?: { code: string; message: string; fields?: Record<string, string> }
}

/** Mirrors apps/server/internal/httpx.APIError — thrown for every non-2xx response. */
export class ApiError extends Error {
  status: number
  code: string
  fields?: Record<string, string>

  constructor(
    status: number,
    code: string,
    message: string,
    fields?: Record<string, string>
  ) {
    super(message)
    this.status = status
    this.code = code
    this.fields = fields
  }
}

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
})

// Every authed call to the Go API carries the JWT from the session store
// (populated once by components/session-provider.tsx) as a bearer token
// — Go verifies it against better-auth's JWKS, see
// apps/server/internal/middlewares/auth.go.
apiClient.interceptors.request.use((config) => {
  const token = useSessionStore.getState().token
  if (token) {
    config.headers.set("Authorization", `Bearer ${token}`)
  }
  return config
})

type RetriableConfig = AxiosRequestConfig & { _retried?: boolean }

apiClient.interceptors.response.use(
  (res) => res,
  async (error: AxiosError) => {
    const original = error.config as RetriableConfig | undefined

    // The bearer JWT is short-lived (15m) by design — a 401 usually just
    // means it expired, not that the session is gone. Mint a fresh one
    // from better-auth's still-valid session cookie and retry once
    // before giving up.
    if (error.response?.status === 401 && original && !original._retried) {
      original._retried = true
      const { data } = await authClient.token()
      if (data?.token) {
        const { organizationId, role } = decodeJwtPayload(data.token)
        useSessionStore.getState().setToken(data.token, organizationId, role)
        original.headers = { ...original.headers, Authorization: `Bearer ${data.token}` }
        return apiClient.request(original)
      }
    }

    if (error.response?.status === 401) {
      useSessionStore.getState().clear()
      if (typeof window !== "undefined") {
        window.location.href = "/login"
      }
    }
    throw error
  }
)

/**
 * The one place network calls happen. Unwraps the {"data"|"error"}
 * envelope every apps/server response uses (see
 * apps/server/internal/httpx/response.go), throws ApiError on failure, and
 * zod-parses the payload on success.
 */
export async function apiRequest<T>(
  config: AxiosRequestConfig,
  schema: ZodType<T, ZodTypeDef, unknown>
): Promise<T> {
  const res = await apiClient.request<ApiEnvelope<T>>(config)
  const body = res.data ?? {}

  if (body.error) {
    throw new ApiError(
      res.status,
      body.error.code ?? "unknown_error",
      body.error.message ?? "Request failed",
      body.error.fields
    )
  }

  return schema.parse(body.data)
}
