"use client";

import { useSessionStore } from "@/store/session.store";

export function useSession() {
    const user = useSessionStore((s) => s.user);
    const token = useSessionStore((s) => s.token);
    const activeOrganizationId = useSessionStore((s) => s.activeOrganizationId);
    const role = useSessionStore((s) => s.role);
    const status = useSessionStore((s) => s.status);
    const clear = useSessionStore((s) => s.clear);
    const setSession = useSessionStore((s) => s.setSession);

    const isAuthenticated = status === "authenticated" && !!token;
    const isLoading = status === "loading" || status === "idle";

    return {
        user,
        token,
        activeOrganizationId,
        role,
        status,
        isAuthenticated,
        isLoading,
        clear,
        setSession,
    };
}

