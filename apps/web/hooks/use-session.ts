"use client";

import { useSessionStore } from "@/store/session.store";

export function useSession() {
    const user = useSessionStore((s) => s.user);
    const token = useSessionStore((s) => s.token);
    const activeOrgId = useSessionStore((s) => s.activeOrgId);
    const isAuthenticated = useSessionStore((s) => s.isAuthenticated);
    const setSession = useSessionStore((s) => s.setSession);
    const setToken = useSessionStore((s) => s.setToken);
    const setUnauthenticated = useSessionStore((s) => s.setUnauthenticated);

    return {
        user,
        token,
        activeOrgId,
        isAuthenticated,
        setSession,
        setToken,
        setUnauthenticated,
    };
}
