"use client";

import * as React from "react";

import { useBreadcrumbStore } from "@/store/breadcrumb.store";

export default function DashboardPage() {
    const setBreadcrumbs = useBreadcrumbStore((state) => state.setBreadcrumbs);

    React.useEffect(() => {
        setBreadcrumbs([
            {
                title: "Dashboard",
            },
        ]);
    }, [setBreadcrumbs]);

    return (
        <div className="flex flex-1 flex-col gap-4">
            <div className="grid auto-rows-min gap-4 md:grid-cols-3">
                <div className="aspect-video rounded-xl bg-muted/50" />
                <div className="aspect-video rounded-xl bg-muted/50" />
                <div className="aspect-video rounded-xl bg-muted/50" />
            </div>
            <div className="min-h-screen flex-1 rounded-xl bg-muted/50 md:min-h-min" />
        </div>
    );
}
