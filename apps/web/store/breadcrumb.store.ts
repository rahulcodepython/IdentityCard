"use client";

import { create } from "zustand";

export interface BreadcrumbItemData {
    label: string;
    href?: string;
}

interface BreadcrumbState {
    items: BreadcrumbItemData[];
    setBreadcrumbs: (items: BreadcrumbItemData[]) => void;
    clearBreadcrumbs: () => void;
}

export const useBreadcrumbStore = create<BreadcrumbState>((set) => ({
    items: [],
    setBreadcrumbs: (items) => set({ items }),
    clearBreadcrumbs: () => set({ items: [] }),
}));