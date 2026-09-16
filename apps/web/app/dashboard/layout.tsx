"use client"

import * as React from "react"
import { AppSidebar } from "@/components/sidebar/app-sidebar"
import BreadcrumbBar from "@/components/sidebar/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
    SidebarInset,
    SidebarProvider,
    SidebarTrigger,
} from "@/components/ui/sidebar"

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode
}) {
    return (
        <SidebarProvider>
            <AppSidebar />
            <SidebarInset>
                <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
                    <div className="flex items-center gap-2 px-4">
                        <SidebarTrigger className="-ml-1" />
                        <Separator
                            orientation="vertical"
                            className="mr-2 data-vertical:h-4 data-vertical:self-auto"
                        />
                        <BreadcrumbBar />
                    </div>
                </header>
                <div className="flex flex-1 flex-col gap-4 p-4 pt-0 w-full max-w-full min-w-0">
                    {children}
                </div>
            </SidebarInset>
        </SidebarProvider>
    )
}
