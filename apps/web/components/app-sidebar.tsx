"use client"

import * as React from "react"
import {
  RiBankCardLine,
  RiBuilding2Line,
  RiCalendarEventLine,
  RiDashboardLine,
  RiSettings3Line,
  RiSmartphoneLine,
  RiUser3Line,
} from "@remixicon/react"

import { NavGroup, NavMain } from "@/components/nav-main"
import { TeamSwitcher } from "@/components/team-switcher"
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"
import type { MeResponse } from "@/lib/validation/auth"

export function AppSidebar({
  user,
  ...props
}: {
  user?: MeResponse
} & React.ComponentProps<typeof Sidebar>) {
  const roles = user?.roles ?? []
  const organizationName = user?.organization_name || "IdentityCard"

  const teams = [
    {
      name: organizationName,
      logo: <RiBuilding2Line className="size-4" />,
      plan: roles.length > 0 ? roles.join(", ") : "Organization",
    },
  ]

  const navGroups: NavGroup[] = [
    {
      items: [
        {
          title: "Dashboard",
          url: "/dashboard",
          icon: <RiDashboardLine className="size-4" />,
        },
      ],
    },
    {
      groupLabel: "Events Management",
      items: [
        {
          title: "Events",
          url: "/dashboard/events",
          icon: <RiCalendarEventLine className="size-4" />,
        },
        {
          title: "Members",
          url: "/dashboard/members",
          icon: <RiUser3Line className="size-4" />,
        },
        ...(roles.includes("super_admin")
          ? [
              {
                title: "Devices",
                url: "/dashboard/devices",
                icon: <RiSmartphoneLine className="size-4" />,
              },
            ]
          : []),
      ],
    },
    {
      groupLabel: "Management",
      items: [
        {
          title: "Billing",
          url: "/dashboard/billing",
          icon: <RiBankCardLine className="size-4" />,
        },
        {
          title: "Settings",
          url: "/dashboard/settings",
          icon: <RiSettings3Line className="size-4" />,
        },
      ],
    },
  ]

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <TeamSwitcher teams={teams} />
      </SidebarHeader>

      <SidebarContent>
        <NavMain groups={navGroups} />
      </SidebarContent>

      <SidebarRail />
    </Sidebar>
  )
}
