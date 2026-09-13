"use client";

import axios, { type AxiosRequestConfig } from "axios";
import type { ZodType, ZodTypeDef } from "zod";

import { API_V1_PREFIX, DEFAULT_API_BASE_URL } from "@/lib/constants";
import { ResponseZod, type Response } from "@/schema/common.types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL;

export class ApiError extends Error {
    constructor(
        public status: number,
        message: string,
    ) {
        super(message);
    }
}

export const apiClient = axios.create({
    baseURL: `${API_BASE_URL}${API_V1_PREFIX}`,
    withCredentials: true,
});

apiClient.interceptors.response.use(
    (res) => res,
    (error) => {
        const status = error.response?.status ?? 500;
        const message =
            error.response?.data?.message ||
            (typeof error.response?.data?.error === "string"
                ? error.response.data.error
                : undefined) ||
            error.message ||
            "Request failed";
        return Promise.reject(new ApiError(status, message));
    },
);

export async function apiRequest<T>(
    config: AxiosRequestConfig,
    schema?: ZodType<T, ZodTypeDef, unknown>,
): Promise<T> {
    const res = await apiClient.request<Response<T>>(config);

    if (schema) {
        const parsed = ResponseZod(schema).parse(res.data);
        return parsed.data as T;
    }

    return res.data.data as T;
}
