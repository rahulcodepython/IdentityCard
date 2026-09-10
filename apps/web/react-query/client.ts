"use client";

import axios, {
    type AxiosError,
    type AxiosRequestConfig,
} from "axios";
import type { ZodType, ZodTypeDef } from "zod";

import { authClient } from "@/lib/auth-client";
import { API_V1_PREFIX } from "@/lib/constants";
import { ApiResponseZod, type ApiResponse } from "@/schema/common.types";
import { useSessionStore } from "@/store/session.store";

const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8000";

export class ApiError extends Error {
    status: number;
    code: string;
    fields?: Record<string, string>;

    constructor(
        status: number,
        code: string,
        message: string,
        fields?: Record<string, string>
    ) {
        super(message);
        this.status = status;
        this.code = code;
        this.fields = fields;
    }
}

export const apiClient = axios.create({
    baseURL: `${API_BASE_URL}${API_V1_PREFIX}`,
});

apiClient.interceptors.request.use((config) => {
    const token = useSessionStore.getState().token;
    if (token) {
        config.headers.set("Authorization", `Bearer ${token}`);
    }
    return config;
});

type RetriableConfig = AxiosRequestConfig & { _retried?: boolean };

let refreshTokenPromise: Promise<string | null> | null = null;

async function getRefreshedToken(): Promise<string | null> {
    if (!refreshTokenPromise) {
        refreshTokenPromise = authClient
            .token()
            .then(({ data }) => {
                if (data?.token) {
                    const activeOrg = useSessionStore.getState().activeOrganizationId;
                    const activeRole = useSessionStore.getState().role;
                    useSessionStore.getState().setToken(data.token, activeOrg, activeRole);
                    return data.token;
                }
                return null;
            })
            .catch(() => null)
            .finally(() => {
                refreshTokenPromise = null;
            });
    }
    return refreshTokenPromise;
}

apiClient.interceptors.response.use(
    (res) => res,
    async (error: AxiosError) => {
        const original = error.config as RetriableConfig | undefined;

        if (error.response?.status === 401 && original && !original._retried) {
            original._retried = true;
            const freshToken = await getRefreshedToken();
            if (freshToken) {
                original.headers = { ...original.headers, Authorization: `Bearer ${freshToken}` };
                return apiClient.request(original);
            }
        }

        if (error.response?.status === 401) {
            useSessionStore.getState().clear();
            if (typeof window !== "undefined" && !window.location.pathname.startsWith("/login")) {
                window.location.href = "/login";
            }
        }

        const data = error.response?.data as ApiResponse<unknown> | undefined;
        let msg = error.message ?? "Request failed";
        let code = "unknown_error";
        let fields: Record<string, string> | undefined;

        if (data) {
            if (typeof data.error === "string") {
                msg = data.error;
            } else if (data.error && typeof data.error === "object") {
                fields = data.error as Record<string, string>;
            }
            if (data.message) {
                code = data.message;
            }
        }

        return Promise.reject(
            new ApiError(
                error.response?.status ?? 500,
                code,
                msg,
                fields
            )
        );
    }
);

export async function apiRequest<T>(
    config: AxiosRequestConfig,
    schema: ZodType<T, ZodTypeDef, unknown>
): Promise<T> {
    const res = await apiClient.request<ApiResponse<T>>(config);
    const envelope = ApiResponseZod(schema).parse(res.data);
    if (!envelope.success) {
        const errDetail = envelope.error;
        let fields: Record<string, string> | undefined;
        if (typeof errDetail === "object" && errDetail !== null) {
            fields = errDetail as Record<string, string>;
        }
        throw new ApiError(
            res.status,
            envelope.message || "Request failed",
            typeof errDetail === "string" ? errDetail : envelope.message || "Request failed",
            fields
        );
    }
    return envelope.data as T;
}
