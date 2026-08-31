"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { RiLoader4Line } from "@remixicon/react";

import { useListSubscriptionsQuery } from "@/query-hooks/plans.api";
import { useSessionStore } from "@/store/session.store";

import { DashboardShell } from "./dashboard-shell";

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const router = useRouter();
    const status = useSessionStore((s) => s.status);
    const user = useSessionStore((s) => s.user);
    const role = useSessionStore((s) => s.role);
    const activeOrganizationId = useSessionStore((s) => s.activeOrganizationId);

    useEffect(() => {
        if (status === "unauthenticated") {
            router.replace("/login");
        }
    }, [status, router]);

    const authenticated = status === "authenticated";
    const subsQuery = useListSubscriptionsQuery(authenticated && !!activeOrganizationId);

    const ready = authenticated && !!user;

    if (!ready) {
        return (
            <div className="flex min-h-svh items-center justify-center">
                <RiLoader4Line className="size-6 animate-spin text-muted-foreground" />
            </div>
        );
    }

    const hasExpiredPlan =
        subsQuery.data?.billings.some(
            (b) => b.status === "pending" || b.status === "cancelled"
        ) ?? false;

    return (
        <DashboardShell user={user} role={role} hasExpiredPlan={hasExpiredPlan}>
            {children}
        </DashboardShell>
    );
}
