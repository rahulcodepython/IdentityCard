import {
  type RemixiconComponentType,
  RiBankCardLine,
  RiCalendarEventLine,
  RiDashboardLine,
  RiSettings3Line,
  RiSmartphoneLine,
  RiUser3Line,
} from "@remixicon/react"

// Single source of truth for the dashboard's nav items — consumed by
// components/app-sidebar.tsx (grouped, icon rendered as JSX) and
// components/dashboard-shell.tsx (flat, for the command palette). A
// member holds exactly one of super_admin/admin/scanner (see
// lib/auth-access-control.ts) — scanner is a real invitable role now,
// not just a device concept, so org-management surfaces (members,
// billing, settings, devices) are admin+ only; scanner still sees the
// dashboard/events overview.
export interface NavItemConfig {
  title: string
  href: string
  icon: RemixiconComponentType
  group?: string
  superAdminOnly?: boolean
  adminOnly?: boolean
}

export const NAV_ITEMS: NavItemConfig[] = [
  { title: "Dashboard", href: "/dashboard", icon: RiDashboardLine },
  {
    title: "Events",
    href: "/dashboard/events",
    icon: RiCalendarEventLine,
    group: "Events Management",
  },
  {
    title: "Members",
    href: "/dashboard/members",
    icon: RiUser3Line,
    group: "Events Management",
    adminOnly: true,
  },
  {
    title: "Devices",
    href: "/dashboard/devices",
    icon: RiSmartphoneLine,
    group: "Events Management",
    superAdminOnly: true,
  },
  { title: "Billing", href: "/dashboard/billing", icon: RiBankCardLine, group: "Management", adminOnly: true },
  { title: "Settings", href: "/dashboard/settings", icon: RiSettings3Line, group: "Management", adminOnly: true },
]

export function getVisibleNavItems(roles: string[]): NavItemConfig[] {
  const isSuperAdmin = roles.includes("super_admin")
  const isAdminOrAbove = isSuperAdmin || roles.includes("admin")
  return NAV_ITEMS.filter((item) => {
    if (item.superAdminOnly && !isSuperAdmin) return false
    if (item.adminOnly && !isAdminOrAbove) return false
    return true
  })
}
