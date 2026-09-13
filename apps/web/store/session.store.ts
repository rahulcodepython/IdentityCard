"use client";

import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

import { STORAGE_KEY_SESSION } from "@/lib/constants";

export interface SessionUser {
    id: string;
    name: string;
    email: string;
    image?: string | null;
}

interface SessionState {
    token: string | null;
    user: SessionUser | null;
    isAuthenticated: boolean;
    activeOrgId: string | null;

    setSession: (s: {
        token: string;
        user: SessionUser;
        activeOrgId?: string | null;
    }) => void;
    setToken: (token: string, activeOrgId?: string | null, role?: string | null) => void;
    setUnauthenticated: () => void;
    changeactiveOrgId: (orgId: string | null) => void;
}

export const useSessionStore = create<SessionState>()(
    persist(
        (set) => ({
            token: null,
            user: null,
            isAuthenticated: false,
            activeOrgId: null,

            setSession: ({ token, user, activeOrgId = null }) =>
                set({
                    token,
                    user,
                    isAuthenticated: !!token,
                    activeOrgId,
                }),

            setToken: (token, activeOrgId = null) =>
                set({
                    token,
                    activeOrgId,
                    isAuthenticated: !!token,
                }),

            setUnauthenticated: () =>
                set({
                    token: null,
                    user: null,
                    isAuthenticated: false,
                    activeOrgId: null,
                }),

            changeactiveOrgId: (orgId: string | null) =>
                set({ activeOrgId: orgId }),
        }),
        {
            name: STORAGE_KEY_SESSION,
            storage: createJSONStorage(() => sessionStorage),
            partialize: (state) => ({
                token: state.token,
                user: state.user,
                isAuthenticated: state.isAuthenticated,
                activeOrgId: state.activeOrgId,
            }),
        }
    )
);

