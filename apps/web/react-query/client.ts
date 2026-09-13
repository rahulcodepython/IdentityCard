"use client";

import axios, { type AxiosError, type AxiosRequestConfig } from "axios";
import type { ZodType, ZodTypeDef } from "zod";

import { authClient } from "@/lib/auth-client";
import {
    API_V1_PREFIX,
    DEFAULT_API_BASE_URL,
    ERR_MSG_REQUEST_FAILED,
    PUBLIC_OR_UNSCOPED_PREFIXES,
    ROUTE_LOGIN,
    TOKEN_EXPIRY_BUFFER_MS,
} from "@/lib/constants";
import { ApiResponseZod, type ApiResponse } from "@/schema/common.types";
import { useSessionStore } from "@/store/session.store";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL;

export class ApiError extends Error {
    status: number;
    code: string;
    fields?: Record<string, string>;

    constructor(
        status: number,
        code: string,
        message: string,
        fields?: Record<string, string>,
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

// Helper: check if a JWT is valid based on its exp claim (with network transit buffer)
function isTokenValid(token: string | null): boolean {
    if (!token) return false;
    try {
        const payload = JSON.parse(atob(token.split(".")[1]));
        return typeof payload.exp === "number" && payload.exp * 1000 > Date.now() + TOKEN_EXPIRY_BUFFER_MS;
    } catch {
        return false;
    }
}

let tokenFetchPromise: Promise<string | null> | null = null;

async function getValidToken(): Promise<string | null> {
    const currentToken = useSessionStore.getState().token;
    if (isTokenValid(currentToken)) {
        return currentToken;
    }

    // Deduplicate simultaneous requests needing a fresh token
    if (!tokenFetchPromise) {
        tokenFetchPromise = authClient
            .token()
            .then(({ data }) => {
                if (data?.token) {
                    const activeOrgId = useSessionStore.getState().activeOrgId;
                    useSessionStore.getState().setToken(data.token, activeOrgId);
                    return data.token;
                }
                return null;
            })
            .catch(() => null)
            .finally(() => {
                tokenFetchPromise = null;
            });
    }

    return tokenFetchPromise;
}

apiClient.interceptors.request.use(async (config) => {
    const token = await getValidToken();
    if (token) {
        config.headers.set("Authorization", `Bearer ${token}`);
    }

    const activeOrgId = useSessionStore.getState().activeOrgId;
    if (activeOrgId && config.url) {
        const isUnscoped = PUBLIC_OR_UNSCOPED_PREFIXES.some((prefix) =>
            config.url?.startsWith(prefix),
        );
        if (!isUnscoped) {
            const path = config.url.startsWith("/") ? config.url : `/${config.url}`;
            config.url = `/organization/${activeOrgId}${path}`;
        }
    }

    return config;
});

apiClient.interceptors.response.use(
    (res) => res,
    (error: AxiosError) => {
        if (error.response?.status === 401) {
            useSessionStore.getState().setUnauthenticated();
            if (
                typeof window !== "undefined" &&
                !window.location.pathname.startsWith(ROUTE_LOGIN)
            ) {
                window.location.href = ROUTE_LOGIN;
            }
        }

        const data = error.response?.data as ApiResponse<unknown> | undefined;
        const msg = data?.message || (typeof data?.error === "string" ? data.error : error.message) || ERR_MSG_REQUEST_FAILED;
        const fields = typeof data?.error === "object" ? (data.error as Record<string, string>) : undefined;

        return Promise.reject(new ApiError(error.response?.status ?? 500, msg, msg, fields));
    },
);

export async function apiRequest<T>(
    config: AxiosRequestConfig,
    schema: ZodType<T, ZodTypeDef, unknown>,
): Promise<T> {
    const res = await apiClient.request<ApiResponse<T>>(config);
    if (res.status === 204 || !res.data || typeof res.data !== "object") {
        return { success: true } as unknown as T;
    }

    const envelope = ApiResponseZod(schema).parse(res.data);
    if (!envelope.success) {
        const errDetail = envelope.error;
        const fields = typeof errDetail === "object" && errDetail !== null ? (errDetail as Record<string, string>) : undefined;
        const msg = envelope.message || (typeof errDetail === "string" ? errDetail : ERR_MSG_REQUEST_FAILED);
        throw new ApiError(res.status, msg, msg, fields);
    }

    return (envelope.data ?? { success: true }) as T;
}
