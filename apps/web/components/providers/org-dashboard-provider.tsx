"use client";

import { useParams } from "next/navigation";
import {
    RiBuilding2Line,
    RiErrorWarningLine,
    RiLoader4Line,
    RiLogoutBoxRLine,
    RiSettings3Line,
    RiTicket2Line,
    RiUser3Line,
} from "@remixicon/react";
import Link from "next/link";
import * as React from "react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import useUser from "@/hooks/use-user";
import useOrganization from "@/hooks/use-organization";
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
    SidebarInset,
    SidebarProvider,
    SidebarTrigger,
} from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/navigation/app-sidebar";
import { useBreadcrumbStore } from "@/store/breadcrumb.store";
import { useBillingOverviewQuery } from "@/query-hooks/plans.api";
import { ModeToggle } from "@/components/ui/theme-toggler";

export default function OrgDashboardProvider({ children }: {
    children: React.ReactNode;
}) {
    const orgSlug = useParams<{ orgSlug: string }>()?.orgSlug;

    const { user, handleLogout } = useUser();
    const { currentOrg, isLoading, isOwner } = useOrganization(orgSlug);

    const items = useBreadcrumbStore((state) => state.items);

    const { data: billing } = useBillingOverviewQuery(isOwner);

    // If still loading organizations or route redirection in progress
    if (isLoading) {
        return (
            <div className="flex min-h-svh items-center justify-center">
                <RiLoader4Line className="size-6 animate-spin text-muted-foreground" />
            </div>
        );
    }

    const dashboardHref = orgSlug ? `/dashboard/${orgSlug}` : "/dashboard";

    return (
        <SidebarProvider>
            {orgSlug && <AppSidebar isOwner={isOwner} orgSlug={orgSlug} />}

            <SidebarInset>
                <header className="flex h-14 shrink-0 items-center justify-between gap-2 border-b px-4">
                    <div className="flex items-center gap-2">
                        {orgSlug && <SidebarTrigger />}
                        <Breadcrumb>
                            <BreadcrumbList>
                                {items.length === 0 ? (
                                    <BreadcrumbItem>
                                        <BreadcrumbPage>Dashboard</BreadcrumbPage>
                                    </BreadcrumbItem>
                                ) : (
                                    <React.Fragment>
                                        <BreadcrumbItem>
                                            <BreadcrumbLink render={<Link href={dashboardHref} />}>
                                                Dashboard
                                            </BreadcrumbLink>
                                        </BreadcrumbItem>
                                        <BreadcrumbSeparator />
                                        {items.map((item, index) => {
                                            const isLast = index === items.length - 1;
                                            return (
                                                <React.Fragment key={item.href || index}>
                                                    <BreadcrumbItem>
                                                        {isLast || !item.href ? (
                                                            <BreadcrumbPage className="max-w-[150px] truncate sm:max-w-[300px]">
                                                                {item.label}
                                                            </BreadcrumbPage>
                                                        ) : (
                                                            <BreadcrumbLink render={<Link href={item.href} />}>
                                                                {item.label}
                                                            </BreadcrumbLink>
                                                        )}
                                                    </BreadcrumbItem>
                                                    {!isLast && <BreadcrumbSeparator />}
                                                </React.Fragment>
                                            );
                                        })}
                                    </React.Fragment>
                                )}
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>

                    <div className="flex items-center gap-2.5">
                        {/* Event Credits Chip (Owner Only) */}
                        {
                            isOwner && billing && (
                                <Link
                                    href={orgSlug ? `/dashboard/${orgSlug}/billing` : "/dashboard/billing"}
                                    className={cn(
                                        "flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium border transition-colors shadow-2xs",
                                        billing.credit_balance === 0
                                            ? "bg-amber-500/10 text-amber-700 border-amber-300 dark:text-amber-300 dark:border-amber-700/50 hover:bg-amber-500/20"
                                            : "bg-primary/5 text-primary border-primary/20 hover:bg-primary/10"
                                    )}
                                >
                                    <RiTicket2Line className="size-3.5" />
                                    <span>{billing.credit_balance} {billing.credit_balance === 1 ? "credit" : "credits"}</span>
                                </Link>
                            )
                        }

                        <ModeToggle />

                        <DropdownMenu>
                            <DropdownMenuTrigger
                                render={
                                    <button className={cn(buttonVariants({ variant: "outline" }), "flex items-center gap-2 px-2.5 py-1 h-8 rounded-lg")}>
                                        <div className="flex size-5 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary font-medium text-[11px]">
                                            {user?.name ? user?.name.charAt(0).toUpperCase() : <RiUser3Line className="size-3.5" />}
                                        </div>
                                        <span className="hidden md:inline-block text-xs font-medium max-w-30 truncate">
                                            {user?.name}
                                        </span>
                                    </button>
                                }
                            />
                            <DropdownMenuContent align="end" className="w-56">
                                <DropdownMenuLabel className="font-normal">
                                    <div className="flex flex-col space-y-1">
                                        <p className="text-sm font-medium leading-none">{user?.name}</p>
                                        <p className="text-xs leading-none text-muted-foreground">{user?.email}</p>
                                    </div>
                                </DropdownMenuLabel>
                                <DropdownMenuSeparator />
                                {
                                    isOwner && (
                                        <React.Fragment>
                                            <DropdownMenuItem render={<Link href={orgSlug ? `/dashboard/${orgSlug}/billing` : "/dashboard/billing"} />}>
                                                <RiTicket2Line className="size-4 mr-2" />
                                                <span>Billing & Credits</span>
                                            </DropdownMenuItem>
                                            <DropdownMenuItem render={<Link href={orgSlug ? `/dashboard/${orgSlug}/settings` : "/dashboard/settings"} />}>
                                                <RiSettings3Line className="size-4 mr-2" />
                                                <span>Settings</span>
                                            </DropdownMenuItem>
                                            <DropdownMenuSeparator />
                                        </React.Fragment>
                                    )
                                }
                                <DropdownMenuItem
                                    onClick={() => handleLogout()}
                                    variant="destructive"
                                >
                                    <RiLogoutBoxRLine className="size-4 mr-2" />
                                    <span>Sign out</span>
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                </header>

                {/* Overdue Annual Maintenance Banner (Owner Only) */}
                {
                    isOwner && billing?.annual_fee_status === "past_due" && (
                        <div className="flex items-center justify-between gap-4 border-b border-destructive/30 bg-destructive/10 px-4 py-2.5 text-xs sm:text-sm text-destructive dark:bg-destructive/20">
                            <div className="flex items-center gap-2">
                                <RiErrorWarningLine className="size-4 shrink-0" />
                                <span>
                                    <strong>Annual Maintenance Overdue:</strong> Your organization retains events and your annual maintenance fee is past due. Event creation is paused until renewal is completed.
                                </span>
                            </div>
                            <Link href={orgSlug ? `/dashboard/${orgSlug}/billing` : "/dashboard/billing"}
                                className="shrink-0 rounded-lg bg-destructive px-3.5 py-1 text-xs font-semibold text-white shadow-2xs hover:bg-destructive/90"
                            >
                                Renew Maintenance
                            </Link>
                        </div>
                    )
                }

                <main className="flex-1 p-6">
                    {children}
                </main>
            </SidebarInset>
        </SidebarProvider>
    );
}
