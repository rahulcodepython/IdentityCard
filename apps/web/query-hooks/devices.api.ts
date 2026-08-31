"use client";

import { apiRequest } from "@/react-query/client";
import {
    useArrayMutation,
    useWithExecute,
} from "@/react-query/mutation";
import { useAppQuery } from "@/react-query/query";
import { queryKeys } from "@/react-query/query-keys";
import {
    type CreateDeviceInput,
    type CreateDeviceResult,
    type Device,
    createDeviceResponseSchema,
    devicesListResponseSchema,
} from "@/schema/devices.types";
import { z } from "zod";

export function useDevicesListQuery(enabled = true) {
    return useAppQuery<Device[]>(
        queryKeys.devices(),
        () =>
            apiRequest(
                { url: "/devices", method: "GET" },
                devicesListResponseSchema
            ),
        { enabled }
    );
}

export function useCreateDeviceMutation() {
    return useWithExecute(
        useArrayMutation<Device, CreateDeviceResult, CreateDeviceInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: "/devices",
                        method: "POST",
                        data,
                    },
                    createDeviceResponseSchema
                ),
            queryKey: queryKeys.devices(),
            updater: (created) => (old) => [
                ...(old ?? []),
                {
                    id: created.id,
                    name: created.name,
                    status: created.status,
                    created_at: created.created_at,
                    verified_at: created.verified_at,
                },
            ],
            showToast: { success: "Device created — enter OTP on device" },
        })
    );
}

export function useRevokeDeviceMutation() {
    return useWithExecute(
        useArrayMutation<Device, void, string>({
            mutationFn: (deviceId) =>
                apiRequest(
                    {
                        url: `/devices/${deviceId}`,
                        method: "DELETE",
                    },
                    z.any()
                ),
            queryKey: queryKeys.devices(),
            updater: (_res, deviceId) => (old) =>
                (old ?? []).map((d) => (d.id === deviceId ? { ...d, status: "revoked" } : d)),
            showToast: { success: "Device access revoked" },
        })
    );
}

