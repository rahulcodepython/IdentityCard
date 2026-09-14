"use client"

import * as React from "react"
import {
    AudioLinesIcon,
    BookOpenIcon,
    BotIcon,
    Calendar,
    FrameIcon,
    GalleryVerticalEndIcon,
    LayoutDashboard,
    MapIcon,
    MapPinIcon,
    PieChartIcon,
    Settings2Icon,
    TerminalIcon,
    TerminalSquareIcon,
} from "lucide-react"

import { NavItems } from "@/components/sidebar/nav-items"
import { NavUser } from "@/components/sidebar/nav-user"
import { TeamSwitcher } from "@/components/sidebar/team-switcher"
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarRail,
} from "@/components/ui/sidebar"
import type { SidebarData } from "@/schema/sidebar.types"

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
    dashboard: {
        label: "Dashboard",
        items: [
            {
                name: "Dashboard",
                url: "/dashboard",
                icon: LayoutDashboard,
            },
        ]
    },
    management: {
        label: "Management",
        items: [
            {
                name: "Events",
                url: "/dashboard/events",
                icon: Calendar,
            },
        ]
    }
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
    return (
        <Sidebar collapsible="icon" {...props}>
            <SidebarHeader>
                <TeamSwitcher teams={data.teams} />
            </SidebarHeader>
            <SidebarContent>
                <NavItems props={data.dashboard} />
                <NavItems props={data.management} />
            </SidebarContent>
            <SidebarFooter>
                <NavUser user={data.user} />
            </SidebarFooter>
            <SidebarRail />
        </Sidebar>
    )
}
