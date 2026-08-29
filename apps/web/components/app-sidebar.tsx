"use client"

import * as React from "react"
import { RiBuilding2Line, RiCheckLine, RiExpandUpDownLine } from "@remixicon/react"

import { NavGroup, NavMain } from "@/components/nav-main"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
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
import { authClient } from "@/lib/auth-client"
import { decodeJwtPayload } from "@/lib/jwt"
import { useSessionStore } from "@/store/session.store"

// The sidebar header is an organization switcher, not a static name
// block — a user can be a member of more than one org (invited into
// others as admin/scanner; they can only ever *create* one themselves,
// see lib/auth.ts's organizationLimit). Switching re-mints the bearer
// token (org id/orgRole are baked into its claims at issue time — see
// apps/server/internal/pkg/jwt), not just a client-side org id swap.
export function AppSidebar({
  orgRole,
  ...props
}: {
  orgRole: string | null
} & React.ComponentProps<typeof Sidebar>) {
  const activeOrganizationId = useSessionStore((s) => s.activeOrganizationId)
  const setToken = useSessionStore((s) => s.setToken)
  const { data: organizations } = authClient.useListOrganizations()
  const [switching, setSwitching] = React.useState(false)

  const activeOrg = organizations?.find((org) => org.id === activeOrganizationId)

  const roles = orgRole ? [orgRole] : []
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

  const switchOrg = async (organizationId: string) => {
    if (organizationId === activeOrganizationId || switching) return
    setSwitching(true)
    const { error } = await authClient.organization.setActive({ organizationId })
    if (!error) {
      const { data } = await authClient.token()
      if (data?.token) {
        const decoded = decodeJwtPayload(data.token)
        setToken(data.token, decoded.organizationId, decoded.role)
      }
    }
    setSwitching(false)
  }

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <SidebarMenuButton
                    size="lg"
                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                  >
                    <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                      <RiBuilding2Line className="size-4" />
                    </div>
                    <div className="grid flex-1 text-left text-sm leading-tight">
                      <span className="truncate font-semibold">
                        {activeOrg?.name ?? "Organization"}
                      </span>
                      <span className="truncate text-xs capitalize">{orgRole?.replace("_", " ") ?? "Member"}</span>
                    </div>
                    <RiExpandUpDownLine className="ml-auto size-4 text-sidebar-foreground/50" />
                  </SidebarMenuButton>
                }
              />
              <DropdownMenuContent align="start" className="w-56">
                {organizations?.length ? (
                  organizations.map((org) => (
                    <DropdownMenuItem key={org.id} onClick={() => void switchOrg(org.id)}>
                      <RiBuilding2Line className="size-4 mr-2" />
                      <span className="flex-1 truncate">{org.name}</span>
                      {org.id === activeOrganizationId && <RiCheckLine className="size-4" />}
                    </DropdownMenuItem>
                  ))
                ) : (
                  <DropdownMenuItem disabled>No organizations</DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
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
