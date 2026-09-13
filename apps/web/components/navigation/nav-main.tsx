"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
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
import { NavGroup } from "@/config/nav"
import { ChevronRight } from "lucide-react"
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from "@/components/ui/collapsible"

export function NavMain({ navGroup }: { navGroup: NavGroup[] }) {
    return (
        navGroup.map((nav) => (
            <SidebarGroup className="group-data-[collapsible=icon]:hidden">
                <SidebarGroupLabel>
                    {nav.label}
                </SidebarGroupLabel>
                <SidebarMenu>
                    {
                        nav.items.map((item) => (
                            (item.url && !item.items) ? <SidebarMenuItem key={item.title}>
                                <SidebarMenuButton>
                                    {
                                        item.url ? <Link href={item.url}>
                                            <item.icon />
                                            <span>{item.title}</span>
                                        </Link> : <span>
                                            <item.icon />
                                            <span>{item.title}</span>
                                        </span>
                                    }
                                </SidebarMenuButton>
                            </SidebarMenuItem> : item.items ? <Collapsible key={item.title} className="group/collapsible">
                                <SidebarMenuItem>
                                    <CollapsibleTrigger>
                                        <SidebarMenuButton tooltip={item.title}>
                                            <item.icon />
                                            <span>{item.title}</span>
                                            <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                        </SidebarMenuButton>
                                    </CollapsibleTrigger>
                                    <CollapsibleContent>
                                        <SidebarMenuSub>
                                            {
                                                item.items.map((subItem) => (
                                                    <SidebarMenuSubItem key={subItem.title}>
                                                        <SidebarMenuSubButton>
                                                            <Link href={subItem.url}>
                                                                <span>{subItem.title}</span>
                                                            </Link>
                                                        </SidebarMenuSubButton>
                                                    </SidebarMenuSubItem>
                                                ))
                                            }
                                        </SidebarMenuSub>
                                    </CollapsibleContent>
                                </SidebarMenuItem>
                            </Collapsible> : null
                        ))
                    }
                </SidebarMenu>
            </SidebarGroup>
        )
        )
    )
}
