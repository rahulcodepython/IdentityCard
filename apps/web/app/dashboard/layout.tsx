"use client";

import { SessionProvider } from "@/components/providers";
import OrgDashboardProvider from "@/components/providers/org-dashboard-provider";

export default function DashboardLayout({ children }: {
    children: React.ReactNode;
}) {
    return (
        <SessionProvider>
            <OrgDashboardProvider>
                {children}
            </OrgDashboardProvider>
        </SessionProvider>
    )
}
