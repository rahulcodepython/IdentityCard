"use client";

import * as React from "react";
import {
    AudioLinesIcon,
    GalleryVerticalEndIcon,
    TerminalIcon,
} from "lucide-react";

import { NavItems } from "./nav-items";
import { NavUser } from "./nav-user";
import { TeamSwitcher } from "./team-switcher";
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarRail,
} from "../ui/sidebar";
import type { SidebarData } from "../../schema/sidebar.types";

// Sample dummy data
const data: SidebarData = {
    user: {
        id: "000000000000000",
        name: "shadcn",
        email: "m@example.com",
        avatar: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&auto=format&fit=crop&q=80",
    },
    teams: [
        {
            name: "Acme Inc",
            logo: GalleryVerticalEndIcon,
            plan: "Enterprise",
        },
        {
            name: "Acme Corp.",
            logo: AudioLinesIcon,
            plan: "Startup",
        },
        {
            name: "Evil Corp.",
            logo: TerminalIcon,
            plan: "Free",
        },
    ],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
    return (
        <Sidebar collapsible="icon" {...props}>
            <SidebarHeader>
                <TeamSwitcher teams={data.teams} />
            </SidebarHeader>
            <SidebarContent>
                <NavItems />
            </SidebarContent>
            <SidebarFooter>
                <NavUser user={data.user} />
            </SidebarFooter>
            <SidebarRail />
        </Sidebar>
    );
}
