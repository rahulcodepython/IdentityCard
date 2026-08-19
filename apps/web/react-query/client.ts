"use client"

import axios, {
  type AxiosError,
  type AxiosRequestConfig,
} from "axios"
import type { ZodType, ZodTypeDef } from "zod"

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
  withCredentials: true, // Send httpOnly session cookies to Go backend
})

apiClient.interceptors.response.use(
  (res) => res,
  (error: AxiosError) => {
    // If the backend returns 401 even after its own auto-refresh attempt,
    // the session is truly dead. Redirect to login.
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
