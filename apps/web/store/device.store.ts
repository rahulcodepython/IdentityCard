"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

const DEVICE_KEY_STORAGE_KEY = "identitycard_device_key";

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
            name: DEVICE_KEY_STORAGE_KEY,
            partialize: (state) => ({ deviceKey: state.deviceKey }),
        }
    )
);
