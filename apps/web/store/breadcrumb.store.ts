"use client";

import { useEffect } from "react";
import { create } from "zustand";

export interface BreadcrumbItem {
    label: string;
    href?: string;
}

interface BreadcrumbState {
    breadcrumbs: BreadcrumbItem[];
    labels: Record<string, string>;
    setBreadcrumbs: (items: BreadcrumbItem[]) => void;
    setLabel: (segmentOrId: string, label: string) => void;
    clearBreadcrumbs: () => void;
}

export const useBreadcrumbStore = create<BreadcrumbState>((set) => ({
    breadcrumbs: [],
    labels: {},
    setBreadcrumbs: (breadcrumbs) => set({ breadcrumbs }),
    setLabel: (segmentOrId, label) =>
        set((state) => ({
            labels: { ...state.labels, [segmentOrId]: label },
        })),
    clearBreadcrumbs: () => set({ breadcrumbs: [], labels: {} }),
}));

export function useSetBreadcrumbs(items: BreadcrumbItem[]) {
    const setBreadcrumbs = useBreadcrumbStore((s) => s.setBreadcrumbs);
    const clearBreadcrumbs = useBreadcrumbStore((s) => s.clearBreadcrumbs);

    useEffect(() => {
        setBreadcrumbs(items);
        return () => clearBreadcrumbs();
    }, [items, setBreadcrumbs, clearBreadcrumbs]);
}

export function useSetBreadcrumbLabel(
    idOrSegment: string | undefined,
    label: string | undefined
) {
    const setLabel = useBreadcrumbStore((s) => s.setLabel);

    useEffect(() => {
        if (idOrSegment && label) {
            setLabel(idOrSegment, label);
        }
    }, [idOrSegment, label, setLabel]);
}

