"use client";

import { useEffect, useMemo } from "react";
import { BreadcrumbItemData, useBreadcrumbStore } from "@/store/breadcrumb.store";

export type { BreadcrumbItemData };

export function useSetBreadcrumbs(items: BreadcrumbItemData[]) {
    const setBreadcrumbs = useBreadcrumbStore((s) => s.setBreadcrumbs);
    const clearBreadcrumbs = useBreadcrumbStore((s) => s.clearBreadcrumbs);

    // Serialized comparison prevents infinite re-render loops when inline array literals are passed
    const serialized = useMemo(() => JSON.stringify(items), [items]);

    useEffect(() => {
        setBreadcrumbs(items);
        return () => {
            clearBreadcrumbs();
        };
    }, [serialized, setBreadcrumbs, clearBreadcrumbs]);
}