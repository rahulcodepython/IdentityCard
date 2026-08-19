"use client"

import { useEffect } from "react"
import { create } from "zustand"

export interface BreadcrumbItem {
  label: string
  href?: string
}

interface BreadcrumbState {
  labels: Record<string, string>
  customBreadcrumbs: BreadcrumbItem[] | null

  setLabel: (segmentOrId: string, label: string) => void
  setLabels: (labelsRecord: Record<string, string>) => void
  setCustomBreadcrumbs: (items: BreadcrumbItem[] | null) => void
  clearBreadcrumbs: () => void
}

export const useBreadcrumbStore = create<BreadcrumbState>((set) => ({
  labels: {},
  customBreadcrumbs: null,

  setLabel: (segmentOrId, label) =>
    set((state) => ({
      labels: { ...state.labels, [segmentOrId]: label },
    })),

  setLabels: (newLabels) =>
    set((state) => ({
      labels: { ...state.labels, ...newLabels },
    })),

  setCustomBreadcrumbs: (items) => set({ customBreadcrumbs: items }),

  clearBreadcrumbs: () => set({ labels: {}, customBreadcrumbs: null }),
}))

export function useSetBreadcrumbLabel(
  idOrSegment: string | undefined,
  label: string | undefined
) {
  const setLabel = useBreadcrumbStore((s) => s.setLabel)

  useEffect(() => {
    if (idOrSegment && label) {
      setLabel(idOrSegment, label)
    }
  }, [idOrSegment, label, setLabel])
}

export function useSetCustomBreadcrumbs(
  items: BreadcrumbItem[] | null
) {
  const setCustomBreadcrumbs = useBreadcrumbStore(
    (s) => s.setCustomBreadcrumbs
  )

  useEffect(() => {
    setCustomBreadcrumbs(items)
    return () => setCustomBreadcrumbs(null)
  }, [items, setCustomBreadcrumbs])
}
