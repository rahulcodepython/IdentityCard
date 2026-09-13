"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import { STORAGE_KEY_DEVICE_KEY } from "@/lib/constants";

interface DeviceState {
    deviceKey: string | null;
    setDeviceKey: (key: string) => void;
    clearDeviceKey: () => void;
}

export const useDeviceStore = create<DeviceState>()(
    persist(
        (set) => ({
            deviceKey: null,
            setDeviceKey: (key) => set({ deviceKey: key }),
            clearDeviceKey: () => set({ deviceKey: null }),
        }),
        {
            name: STORAGE_KEY_DEVICE_KEY,
            partialize: (state) => ({ deviceKey: state.deviceKey }),
        }
    )
);
