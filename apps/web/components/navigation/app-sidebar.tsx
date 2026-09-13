"use client";

import * as React from "react";
import { RiBuilding2Line, RiCheckLine, RiExpandUpDownLine } from "@remixicon/react";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
    Sidebar,
    SidebarContent,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarRail,
} from "@/components/ui/sidebar";
import getNavItems from "@/config/nav";
import useOrganization from "@/hooks/use-organization";
import { NavMain } from "@/components/navigation/nav-main";

export function AppSidebar({ isOwner, orgSlug, ...props }: {
    isOwner: boolean;
    orgSlug: string;
} & React.ComponentProps<typeof Sidebar>) {
    const { currentOrg, organizations, handleSelectOrg } = useOrganization(orgSlug);

    const navGroups = getNavItems(isOwner, orgSlug);

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
                                                {currentOrg?.name ?? "Organization"}
                                            </span>
                                            <span className="truncate text-xs capitalize">
                                                {isOwner ? "Owner" : "Member"}
                                            </span>
                                        </div>
                                        <RiExpandUpDownLine className="ml-auto size-4 text-sidebar-foreground/50" />
                                    </SidebarMenuButton>
                                }
                            />
                            <DropdownMenuContent align="start" className="w-56">
                                {
                                    organizations?.length ? organizations.map((org) => (
                                        <DropdownMenuItem key={org.id} onClick={() => handleSelectOrg(org.slug)}>
                                            <RiBuilding2Line className="size-4 mr-2" />
                                            <span className="flex-1 truncate">{org.name}</span>
                                            {org.id === currentOrg?.id && <RiCheckLine className="size-4" />}
                                        </DropdownMenuItem>
                                    )) : <DropdownMenuItem disabled>No organizations</DropdownMenuItem>

                                }
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>

            <SidebarContent>
                <NavMain navGroup={navGroups} />
            </SidebarContent>

            <SidebarRail />
        </Sidebar>
    );
}
