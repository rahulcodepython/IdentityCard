"use client"

import * as React from "react"
import { ChevronRightIcon } from "lucide-react"

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
import type { NavNestedItems } from "@/schema/sidebar.types"

export function NavNestedItems({ props }: { props: NavNestedItems }) {
    return (
        <SidebarGroup>
            <SidebarGroupLabel>
                {props.label}
            </SidebarGroupLabel>
            <SidebarMenu>
                {
                    props.items.map((item) => (
                        <Collapsible
                            key={item.title}
                            // defaultOpen={item.isActive}
                            className="group/collapsible"
                            render={<SidebarMenuItem />}
                        >
                            <CollapsibleTrigger render={
                                <SidebarMenuButton tooltip={item.title} />
                            }>
                                {item.icon && <item.icon />}
                                <span>{item.title}</span>
                                <ChevronRightIcon className="ml-auto transition-transform duration-200 group-data-open/collapsible:rotate-90" />
                            </CollapsibleTrigger>
                            <CollapsibleContent>
                                <SidebarMenuSub>
                                    {
                                        item.items?.map((subItem) => (
                                            <SidebarMenuSubItem key={subItem.title}>
                                                <SidebarMenuSubButton render={<a href={subItem.url} />}>
                                                    <span>{subItem.title}</span>
                                                </SidebarMenuSubButton>
                                            </SidebarMenuSubItem>
                                        ))
                                    }
                                </SidebarMenuSub>
                            </CollapsibleContent>
                        </Collapsible>
                    ))
                }
            </SidebarMenu>
        </SidebarGroup>
    )
}
