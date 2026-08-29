import {
    type RemixiconComponentType,
    RiBankCardLine,
    RiCalendarEventLine,
    RiDashboardLine,
    RiSettings3Line,
    RiSmartphoneLine,
    RiUser3Line,
} from "@remixicon/react"

export interface NavItemConfig {
    title: string
    href: string
    icon: RemixiconComponentType
    group?: string
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
    },
    {
        title: "Devices",
        href: "/dashboard/devices",
        icon: RiSmartphoneLine,
        group: "Events Management",
    },
    { title: "Billing", href: "/dashboard/billing", icon: RiBankCardLine, group: "Management" },
    { title: "Settings", href: "/dashboard/settings", icon: RiSettings3Line, group: "Management" },
]

export function getVisibleNavItems(roles: string[]): NavItemConfig[] {
    const isAdmin = roles.includes("admin")
    return NAV_ITEMS.filter((item) => {
        if (item.adminOnly && !isAdmin) return false
        return true
    })
}
