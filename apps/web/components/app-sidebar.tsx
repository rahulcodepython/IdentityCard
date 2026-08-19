"use client"

import * as React from "react"
import { RiBuilding2Line } from "@remixicon/react"

import { NavGroup, NavMain } from "@/components/nav-main"
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar"
import { getVisibleNavItems } from "@/config/nav"
import type { MeResponse } from "@/schema/auth.types"

export function AppSidebar({
  user,
  ...props
}: {
  user?: MeResponse
} & React.ComponentProps<typeof Sidebar>) {
  const roles = user?.roles ?? []
  const organizationName = user?.organization_name || "IdentityCard"

  const visibleItems = getVisibleNavItems(roles)
  const ungrouped = visibleItems.filter((item) => !item.group)
  const groupLabels = [...new Set(visibleItems.filter((item) => item.group).map((item) => item.group!))]

  const navGroups: NavGroup[] = [
    {
      items: ungrouped.map((item) => ({
        title: item.title,
        url: item.href,
        icon: <item.icon className="size-4" />,
      })),
    },
    ...groupLabels.map((groupLabel) => ({
      groupLabel,
      items: visibleItems
        .filter((item) => item.group === groupLabel)
        .map((item) => ({
          title: item.title,
          url: item.href,
          icon: <item.icon className="size-4" />,
        })),
    })),
  ]

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground">
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                <RiBuilding2Line className="size-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">
                  {organizationName}
                </span>
                <span className="truncate text-xs">{roles.length > 0 ? roles.join(", ") : "Organization"}</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <NavMain groups={navGroups} />
      </SidebarContent>

      <SidebarRail />
    </Sidebar>
  )
}
