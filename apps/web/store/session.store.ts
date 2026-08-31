"use client";

import { create } from "zustand";

export type SessionStatus = "idle" | "loading" | "authenticated" | "unauthenticated";

export interface SessionUser {
    id: string;
    name: string;
    email: string;
    image?: string | null;
}

interface SessionState {
    token: string | null;
    user: SessionUser | null;
    activeOrganizationId: string | null;
    role: string | null;
    status: SessionStatus;

    setLoading: () => void;
    setSession: (s: {
        token: string;
        user: SessionUser;
        activeOrganizationId: string | null;
        role: string | null;
    }) => void;
    setToken: (token: string, activeOrganizationId: string | null, role: string | null) => void;
    setUnauthenticated: () => void;
    clear: () => void;
}

export const useSessionStore = create<SessionState>((set) => ({
    token: null,
    user: null,
    activeOrganizationId: null,
    role: null,
    status: "idle",

    setLoading: () => set({ status: "loading" }),

    setSession: ({ token, user, activeOrganizationId, role }) =>
        set({ token, user, activeOrganizationId, role, status: "authenticated" }),

    setToken: (token, activeOrganizationId, role) =>
        set({ token, activeOrganizationId, role }),

    setUnauthenticated: () =>
        set({ token: null, user: null, activeOrganizationId: null, role: null, status: "unauthenticated" }),

    clear: () =>
        set({ token: null, user: null, activeOrganizationId: null, role: null, status: "unauthenticated" }),
}));
