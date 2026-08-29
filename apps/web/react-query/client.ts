"use client"

import axios, {
    type AxiosError,
    type AxiosRequestConfig,
} from "axios"
import type { ZodType, ZodTypeDef } from "zod"

import { authClient } from "@/lib/auth-client"
import { API_V1_PREFIX } from "@/lib/constants"
import { decodeJwtPayload } from "@/lib/jwt"
import { useSessionStore } from "@/store/session.store"

const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080"

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

export const apiClient = axios.create({
    baseURL: `${API_BASE_URL}${API_V1_PREFIX}`,
})

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
