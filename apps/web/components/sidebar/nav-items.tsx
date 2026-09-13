"use client"

import {
    SidebarGroup,
    SidebarGroupLabel,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from "@/components/ui/sidebar"
import type { NavItems } from "@/schema/sidebar.types"

export function NavItems({ props }: { props: NavItems }) {

    return (
        <SidebarGroup className="group-data-[collapsible=icon]:hidden">
            <SidebarGroupLabel>
                {props.label}
            </SidebarGroupLabel>
            <SidebarMenu>
                {
                    props.items.map((item) => (
                        <SidebarMenuItem key={item.name}>
                            <SidebarMenuButton render={<a href={item.url} />}>
                                <item.icon className="-mt-1" />
                                <span>{item.name}</span>
                            </SidebarMenuButton>
                        </SidebarMenuItem>
                    ))
                }
            </SidebarMenu>
        </SidebarGroup>
    )
}
