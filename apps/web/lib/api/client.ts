import "server-only"
import { headers } from "next/headers"

import { auth } from "@/lib/auth"
import { API_V1_PREFIX } from "@/lib/constants"

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

type ApiEnvelope<T> = {
    data?: T
    error?: { code: string; message: string; fields?: Record<string, string> }
}

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

export async function apiFetch<T = unknown>(
    path: string,
    init: RequestInit = {}
): Promise<T> {
    const reqHeaders = await headers()
    const { token } = await auth.api.getToken({ headers: reqHeaders })

    const isFormData = init.body instanceof FormData

    const res = await fetch(`${API_BASE_URL}${API_V1_PREFIX}${path}`, {
        ...init,
        headers: {
            ...(isFormData ? {} : { "Content-Type": "application/json" }),
            Authorization: `Bearer ${token}`,
            ...init.headers,
        },
        cache: "no-store",
    })

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
