"use client"

import { z } from "zod"

import { apiRequest } from "@/react-query/client"
import { useSimpleMutation, useWithExecute } from "@/react-query/mutation"
import { useAppQuery } from "@/react-query/query"
import { queryKeys } from "@/react-query/query-keys"
import {
  meResponseSchema,
  type MeResponse,
  messageResponseSchema,
  type RegisterInput,
  registerResponseSchema,
  type SendOtpInput,
  type VerifyOtpInput,
  type VerifyTotpInput,
} from "@/schema/auth.types"
import { useSessionStore } from "@/store/session.store"

// These six hooks call the LOCAL Next.js proxy routes under app/api/auth/*
// (same-origin, plain fetch) — those are the only place the httpOnly
// session cookie is readable/writable, see lib/auth-cookies.server.ts.
// useMeQuery below is the odd one out: it calls Go directly via
// react-query/client.ts's apiRequest, with the Bearer token the interceptor
// attaches from store/session.store.ts.

const accessTokenResultSchema = z.object({ accessToken: z.string() })

async function callProxy(path: string, payload: unknown = {}): Promise<unknown> {
  const res = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
  const body = (await res.json().catch(() => ({}))) as {
    data?: unknown
    error?: { code: string; message: string }
  }
  if (!res.ok || body?.error) {
    throw new Error(body?.error?.message ?? "Request failed")
  }
  return body
}

export function useSendOtpMutation() {
  return useWithExecute(
    useSimpleMutation<{ message: string }, SendOtpInput>({
      mutationFn: async (input) => {
        const body = (await callProxy("/api/auth/otp/send", input)) as { data: unknown }
        return messageResponseSchema.parse(body.data)
      },
    })
  )
}

export function useVerifyOtpMutation() {
  const setSession = useSessionStore((s) => s.setSession)
  return useWithExecute(
    useSimpleMutation<{ accessToken: string }, VerifyOtpInput>({
      mutationFn: async (input) => {
        const body = await callProxy("/api/auth/otp/verify", input)
        return accessTokenResultSchema.parse(body)
      },
      onSuccessExtra: (data) => setSession({ accessToken: data.accessToken }),
    })
  )
}

export function useVerifyTotpMutation() {
  const setSession = useSessionStore((s) => s.setSession)
  return useWithExecute(
    useSimpleMutation<{ accessToken: string }, VerifyTotpInput>({
      mutationFn: async (input) => {
        const body = await callProxy("/api/auth/totp/verify", input)
        return accessTokenResultSchema.parse(body)
      },
      onSuccessExtra: (data) => setSession({ accessToken: data.accessToken }),
    })
  )
}

export function useRegisterMutation() {
  return useWithExecute(
    useSimpleMutation<
      { email: string; totp_qr_image: string; totp_secret: string },
      RegisterInput
    >({
      mutationFn: async (input) => {
        const body = (await callProxy("/api/auth/register", input)) as { data: unknown }
        return registerResponseSchema.parse(body.data)
      },
    })
  )
}

export function useCreateOrganizationMutation() {
  const setSession = useSessionStore((s) => s.setSession)
  return useWithExecute(
    useSimpleMutation<{ accessToken: string }, { organization_name: string }>({
      mutationFn: async (input) => {
        const body = await callProxy("/api/auth/organization", input)
        return accessTokenResultSchema.parse(body)
      },
      onSuccessExtra: (data) => setSession({ accessToken: data.accessToken }),
    })
  )
}

export function useLogoutMutation() {
  return useWithExecute(
    useSimpleMutation<{ message: string }, void>({
      mutationFn: async () => {
        const body = await callProxy("/api/auth/logout", {})
        return messageResponseSchema.parse(body)
      },
    })
  )
}

export function useMeQuery(enabled = true) {
  return useAppQuery<MeResponse>(
    queryKeys.me(),
    () => apiRequest({ url: "/auth/me", method: "GET" }, meResponseSchema),
    { enabled, retry: false }
  )
}
