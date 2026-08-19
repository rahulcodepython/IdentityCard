"use client"

import { useSetBreadcrumbLabel } from "@/lib/stores/use-breadcrumb-store"

export function BreadcrumbSetter({ id, label }: { id: string; label: string }) {
  useSetBreadcrumbLabel(id, label)
  return null
}
