"use client"

import { create } from "zustand"

import type { MeResponse } from "@/schema/auth.types"

export type SessionStatus = "idle" | "loading" | "authenticated" | "unauthenticated"

interface SessionState {
  accessToken: string | null
  user: MeResponse | null
  status: SessionStatus

  setSession: (session: { accessToken: string; user?: MeResponse | null }) => void
  setUser: (user: MeResponse | null) => void
  clear: () => void
  hydrate: () => Promise<void>
}

// accessToken is deliberately never persisted (no zustand `persist` here) —
// it lives in memory only. Page-load rehydration goes through GET
// /api/session, which reads the httpOnly cookie server-side (see
// app/api/session/route.ts) and hands the token back to this store.
export const useSessionStore = create<SessionState>((set, get) => ({
  accessToken: null,
  user: null,
  status: "idle",

  setSession: ({ accessToken, user }) =>
    set((state) => ({
      accessToken,
      user: user !== undefined ? user : state.user,
      status: "authenticated",
    })),

  setUser: (user) => set({ user }),

  clear: () => set({ accessToken: null, user: null, status: "unauthenticated" }),

  hydrate: async () => {
    if (get().status === "loading") return
    set({ status: "loading" })
    try {
      const res = await fetch("/api/session", { method: "GET" })
      if (res.status === 200) {
        const body = (await res.json()) as { accessToken: string }
        set({ accessToken: body.accessToken, status: "authenticated" })
        return
      }
      set({ accessToken: null, user: null, status: "unauthenticated" })
    } catch {
      set({ accessToken: null, user: null, status: "unauthenticated" })
    }
  },
}))
