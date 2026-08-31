"use client";

import {
    RiCalendarEventLine,
    RiComputerLine,
    RiErrorWarningLine,
    RiLogoutBoxRLine,
    RiMoonLine,
    RiSettings3Line,
    RiSunLine,
    RiUser3Line,
} from "@remixicon/react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import * as React from "react";
import { useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import {
    CommandDialog,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
} from "@/components/ui/command";
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
import { useBreadcrumbStore } from "@/store/breadcrumb.store";
import { getVisibleNavItems } from "@/config/nav";
import { authClient } from "@/lib/auth-client";
import { useSessionStore, type SessionUser } from "@/store/session.store";

import { AppSidebar } from "@/components/app-sidebar";

export function DashboardShell({
    user,
    role,
    hasExpiredPlan = false,
    children,
}: {
    user: SessionUser;
    role: string | null;
    hasExpiredPlan?: boolean;
    children: React.ReactNode;
}) {
    const pathname = usePathname();
    const router = useRouter();
    const { setTheme } = useTheme();
    const [commandOpen, setCommandOpen] = useState(false);
    const [isPending, setIsPending] = useState(false);
    const clearSession = useSessionStore((s) => s.clear);

    const storeBreadcrumbs = useBreadcrumbStore((s) => s.breadcrumbs);
    const storeLabels = useBreadcrumbStore((s) => s.labels);

    const roles = useMemo(() => (role ? [role] : []), [role]);

    const allNavItems = getVisibleNavItems(roles).map((item) => ({
        href: item.href,
        label: item.title,
        icon: item.icon,
    }));

    const handleLogout = async () => {
        setIsPending(true);
        await authClient.signOut();
        clearSession();
        router.push("/login");
    };

    useEffect(() => {
        function onKeyDown(e: KeyboardEvent) {
            if ((e.metaKey || e.ctrlKey) && e.key === "k") {
                e.preventDefault();
                setCommandOpen((open) => !open);
            }
        }
        document.addEventListener("keydown", onKeyDown);
        return () => document.removeEventListener("keydown", onKeyDown);
    }, []);

    const computedBreadcrumbs = useMemo(() => {
        if (storeBreadcrumbs && storeBreadcrumbs.length > 0) {
            return storeBreadcrumbs;
        }

        const segments = pathname.split("/").filter(Boolean);
        const items: Array<{ label: string; href: string }> = [];
        let currentPath = "";

        for (let i = 0; i < segments.length; i++) {
            const seg = segments[i];
            currentPath += `/${seg}`;

            let label = storeLabels[seg];
            if (!label) {
                if (seg === "dashboard") label = "Dashboard";
                else if (seg === "events") label = "Events";
                else if (seg === "devices") label = "Devices";
                else if (seg === "billing") label = "Billing";
                else if (seg === "settings") label = "Settings";
                else if (seg === "members") label = "Members";
                else if (seg === "people") label = "People";
                else if (seg === "forms") label = "Forms";
                else if (seg === "subevents") label = "Sub-Events";
                else if (seg === "analytics") label = "Analytics";
                else if (seg === "edit") label = "Edit";
                else if (seg === "new") label = "New";
                else if (seg === "import") label = "Import";
                else {
                    label = seg.length > 20 ? `${seg.slice(0, 8)}…` : seg.charAt(0).toUpperCase() + seg.slice(1);
                }
            }

            items.push({ label, href: currentPath });
        }

        return items;
    }, [pathname, storeLabels, storeBreadcrumbs]);

    return (
        <SidebarProvider>
            <AppSidebar orgRole={role} />

            <SidebarInset>
                <header className="flex h-14 shrink-0 items-center justify-between gap-2 border-b px-4">
                    <div className="flex items-center gap-2">
                        <SidebarTrigger />
                        <Breadcrumb>
                            <BreadcrumbList>
                                {computedBreadcrumbs.map((item, index) => {
                                    const isLast = index === computedBreadcrumbs.length - 1;
                                    return (
                                        <React.Fragment key={item.href || index}>
                                            {index > 0 && <BreadcrumbSeparator />}
                                            <BreadcrumbItem>
                                                {isLast ? (
                                                    <BreadcrumbPage>{item.label}</BreadcrumbPage>
                                                ) : (
                                                    <BreadcrumbLink render={<Link href={item.href ?? "#"} />}>
                                                        {item.label}
                                                    </BreadcrumbLink>
                                                )}
                                            </BreadcrumbItem>
                                        </React.Fragment>
                                    );
                                })}
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>

                    <div className="flex items-center gap-2">
                        <DropdownMenu>
                            <DropdownMenuTrigger
                                render={
                                    <button className={cn(buttonVariants({ variant: "outline", size: "icon" }), "size-8")}>
                                        <RiSunLine className="size-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
                                        <RiMoonLine className="absolute size-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
                                        <span className="sr-only">Toggle theme</span>
                                    </button>
                                }
                            />
                            <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => setTheme("light")}>
                                    <RiSunLine className="size-4 mr-2" />
                                    <span>Light</span>
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => setTheme("dark")}>
                                    <RiMoonLine className="size-4 mr-2" />
                                    <span>Dark</span>
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => setTheme("system")}>
                                    <RiComputerLine className="size-4 mr-2" />
                                    <span>System</span>
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>

                        <DropdownMenu>
                            <DropdownMenuTrigger
                                render={
                                    <button className={cn(buttonVariants({ variant: "outline" }), "flex items-center gap-2 px-2.5 py-1 h-8 rounded-lg")}>
                                        <div className="flex size-5 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary font-medium text-[11px]">
                                            {user.name ? user.name.charAt(0).toUpperCase() : <RiUser3Line className="size-3.5" />}
                                        </div>
                                        <span className="hidden md:inline-block text-xs font-medium max-w-30 truncate">
                                            {user.name}
                                        </span>
                                    </button>
                                }
                            />
                            <DropdownMenuContent align="end" className="w-56">
                                <DropdownMenuLabel className="font-normal">
                                    <div className="flex flex-col space-y-1">
                                        <p className="text-sm font-medium leading-none">{user.name}</p>
                                        <p className="text-xs leading-none text-muted-foreground">{user.email}</p>
                                    </div>
                                </DropdownMenuLabel>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem render={<Link href="/dashboard/settings" />}>
                                    <RiSettings3Line className="size-4 mr-2" />
                                    <span>Settings</span>
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    disabled={isPending}
                                    onClick={() => void handleLogout()}
                                    variant="destructive"
                                >
                                    <RiLogoutBoxRLine className="size-4 mr-2" />
                                    <span>{isPending ? "Signing out…" : "Sign out"}</span>
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                </header>

                {hasExpiredPlan && (
                    <div className="flex items-center justify-between gap-4 border-b border-destructive/30 bg-destructive/10 px-4 py-2.5 text-xs sm:text-sm text-destructive dark:bg-destructive/20">
                        <div className="flex items-center gap-2">
                            <RiErrorWarningLine className="size-4 shrink-0" />
                            <span>
                                <strong>Subscription Expired:</strong> Your plan has expired or payment is overdue. Please renew your plan to continue managing events.
                            </span>
                        </div>
                        <Link
                            href="/dashboard/billing"
                            className="shrink-0 rounded-lg bg-destructive px-3.5 py-1 text-xs font-semibold text-black shadow-2xs"
                        >
                            Renew Plan
                        </Link>
                    </div>
                )}

                <main className="flex-1 p-6">{children}</main>
            </SidebarInset>

            <CommandDialog open={commandOpen} onOpenChange={setCommandOpen}>
                <CommandInput placeholder="Jump to..." />
                <CommandList>
                    <CommandEmpty>No results found.</CommandEmpty>
                    <CommandGroup heading="Navigate">
                        {allNavItems.map((item) => (
                            <CommandItem
                                key={item.href}
                                onSelect={() => {
                                    setCommandOpen(false);
                                    router.push(item.href);
                                }}
                            >
                                <item.icon />
                                <span>{item.label}</span>
                            </CommandItem>
                        ))}
                    </CommandGroup>
                    <CommandGroup heading="Actions">
                        <CommandItem
                            onSelect={() => {
                                setCommandOpen(false);
                                router.push("/dashboard/events?new=1");
                            }}
                        >
                            <RiCalendarEventLine />
                            <span>New event</span>
                        </CommandItem>
                    </CommandGroup>
                </CommandList>
            </CommandDialog>
        </SidebarProvider>
    );
}
