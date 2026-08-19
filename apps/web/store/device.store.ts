"use client"

import { create } from "zustand"
import { persist } from "zustand/middleware"

// Replaces the raw localStorage calls in lib/device-client.ts with the same
// storage key, so an already-paired device stays paired across this
// migration. A scanner-bot device never logs in as an org user — no JWT, no
// cookie, see PROJECT_MEMORY.md's "one exception" note — so this is
// persisted (unlike session.store.ts's in-memory accessToken) on purpose.
const DEVICE_KEY_STORAGE_KEY = "identitycard_device_key"

interface DeviceState {
  deviceKey: string | null
  setDeviceKey: (key: string) => void
  clearDeviceKey: () => void
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
)
