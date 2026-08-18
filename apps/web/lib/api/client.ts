import "server-only"
import { cookies } from "next/headers"

import { forwardSetCookies } from "@/lib/api/cookies"

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

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

/**
 * Typed fetch wrapper for calling apps/server. Forwards the request's auth
 * cookies to the API and re-applies any `Set-Cookie` it returns. Every
 * caller in lib/api/<module>.ts still parses the returned data through its
 * own zod schema — this only guarantees the transport envelope is honored.
 */
export async function apiFetch<T = unknown>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const cookieStore = await cookies()

  // A FormData body (CSV import) needs fetch to set its own multipart
  // boundary — force-setting Content-Type here would break it.
  const isFormData = init.body instanceof FormData

  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      ...(isFormData ? {} : { "Content-Type": "application/json" }),
      Cookie: cookieStore.toString(),
      ...init.headers,
    },
    cache: "no-store",
  })

  await forwardSetCookies(res)

  const body = (await res.json().catch(() => ({}))) as ApiEnvelope<T>

  if (!res.ok || body.error) {
    throw new ApiError(
      res.status,
      body.error?.code ?? "unknown_error",
      body.error?.message ?? "Request failed",
      body.error?.fields
    )
  }

  return body.data as T
}
