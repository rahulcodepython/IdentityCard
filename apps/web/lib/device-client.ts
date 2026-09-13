"use client";

import {
    API_V1_PREFIX,
    DEFAULT_API_BASE_URL,
    ERR_MSG_REQUEST_FAILED,
    STORAGE_KEY_DEVICE_KEY,
} from "@/lib/constants";
import {
    pairResponseSchema,
    scannerMeResponseSchema,
    scanResponseSchema,
} from "@/schema/scanner.types";

const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL;

export function getDeviceKey(): string | null {
    if (typeof window === "undefined") return null;
    return window.localStorage.getItem(STORAGE_KEY_DEVICE_KEY);
}

export function setDeviceKey(key: string) {
    window.localStorage.setItem(STORAGE_KEY_DEVICE_KEY, key);
}

export function clearDeviceKey() {
    window.localStorage.removeItem(STORAGE_KEY_DEVICE_KEY);
}

export class DeviceApiError extends Error {
    status: number;
    code: string;
    constructor(status: number, code: string, message: string) {
        super(message);
        this.status = status;
        this.code = code;
    }
}

async function deviceFetch<T>(
    path: string,
    init: RequestInit = {}
): Promise<T> {
    const key = getDeviceKey();
    const res = await fetch(`${API_BASE_URL}${API_V1_PREFIX}${path}`, {
        ...init,
        headers: {
            "Content-Type": "application/json",
            ...(key ? { "X-Device-Key": key } : {}),
            ...init.headers,
        },
    });

    const body = await res.json().catch(() => ({}));
    if (!res.ok || body.error) {
        throw new DeviceApiError(
            res.status,
            body.error?.code ?? "unknown_error",
            body.error?.message ?? ERR_MSG_REQUEST_FAILED
        );
    }
    return body.data as T;
}

export async function pairDevice(otpCode: string) {
    const data = await deviceFetch("/public/devices/pair", {
        method: "POST",
        body: JSON.stringify({ otp_code: otpCode }),
    });
    return pairResponseSchema.parse(data);
}

export async function getScannerMe() {
    const data = await deviceFetch("/scanner/me");
    return scannerMeResponseSchema.parse(data);
}

export async function scanQr(qrToken: string) {
    const data = await deviceFetch("/scanner/scan", {
        method: "POST",
        body: JSON.stringify({ qr_token: qrToken }),
    });
    return scanResponseSchema.parse(data);
}
