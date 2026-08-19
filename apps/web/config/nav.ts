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
// components/dashboard-shell.tsx (flat, for the command palette). Only one
// role exists today (dashboard org member), but the super_admin gate for
// "Devices" no longer lives in two places.
export interface NavItemConfig {
  title: string
  href: string
  icon: RemixiconComponentType
  group?: string
  superAdminOnly?: boolean
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
  },
  {
    title: "Devices",
    href: "/dashboard/devices",
    icon: RiSmartphoneLine,
    group: "Events Management",
    superAdminOnly: true,
  },
  { title: "Billing", href: "/dashboard/billing", icon: RiBankCardLine, group: "Management" },
  { title: "Settings", href: "/dashboard/settings", icon: RiSettings3Line, group: "Management" },
]

export function getVisibleNavItems(roles: string[]): NavItemConfig[] {
  return NAV_ITEMS.filter((item) => !item.superAdminOnly || roles.includes("super_admin"))
}
