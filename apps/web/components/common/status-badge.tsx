import { Badge, badgeVariants } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { VariantProps } from "class-variance-authority"

type BadgeVariant = VariantProps<typeof badgeVariants>["variant"]

export interface StatusBadgeMapEntry {
  label: string
  variant?: BadgeVariant
  className?: string
}

// Data-driven status badge (template §1.6) — replaces N hand-rolled
// if/else badge components (one per entity's status enum) with one
// lookup-driven component.
export interface StatusBadgeProps {
  status: string
  map: Record<string, StatusBadgeMapEntry>
  fallback?: StatusBadgeMapEntry
}

export function StatusBadge({ status, map, fallback }: StatusBadgeProps) {
  const entry = map[status] ?? fallback ?? { label: status, variant: "outline" as const }
  return (
    <Badge variant={entry.variant} className={cn(entry.className)}>
      {entry.label}
    </Badge>
  )
}
