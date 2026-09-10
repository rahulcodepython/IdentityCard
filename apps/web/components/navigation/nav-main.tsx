"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { RiArrowRightSLine } from "@remixicon/react"

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar"

export type NavItem = {
  title: string
  url: string
  icon?: React.ReactNode
  isActive?: boolean
  items?: {
    title: string
    url: string
  }[]
}

export type NavGroup = {
  groupLabel?: string
  items: NavItem[]
}

export function NavMain({ groups }: { groups: NavGroup[] }) {
  const pathname = usePathname()

  return (
    <>
      {groups.map((group, groupIdx) => (
        <SidebarGroup key={group.groupLabel || groupIdx}>
          {group.groupLabel && (
            <SidebarGroupLabel>{group.groupLabel}</SidebarGroupLabel>
          )}
          <SidebarMenu>
            {group.items.map((item) => {
              const isItemActive =
                pathname === item.url ||
                (item.url !== "/dashboard" && pathname.startsWith(item.url))

              if (item.items && item.items.length > 0) {
                return (
                  <Collapsible
                    key={item.title}
                    defaultOpen={item.isActive || isItemActive}
                    className="group/collapsible"
                    render={<SidebarMenuItem />}
                  >
                    <CollapsibleTrigger
                      render={<SidebarMenuButton tooltip={item.title} isActive={isItemActive} />}
                    >
                      {item.icon}
                      <span>{item.title}</span>
                      <RiArrowRightSLine className="ml-auto transition-transform duration-200 group-data-open/collapsible:rotate-90" />
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                      <SidebarMenuSub>
                        {item.items.map((subItem) => (
                          <SidebarMenuSubItem key={subItem.title}>
                            <SidebarMenuSubButton
                              isActive={pathname === subItem.url}
                              render={<Link href={subItem.url} />}
                            >
                              <span>{subItem.title}</span>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
                        ))}
                      </SidebarMenuSub>
                    </CollapsibleContent>
                  </Collapsible>
                )
              }

              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    isActive={isItemActive}
                    tooltip={item.title}
                    render={<Link href={item.url} />}
                  >
                    {item.icon}
                    <span>{item.title}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )
            })}
          </SidebarMenu>
        </SidebarGroup>
      ))}
    </>
  )
}
