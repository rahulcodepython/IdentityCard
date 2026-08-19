"use client"

import Link from "next/link"
import { useQuery } from "@tanstack/react-query"
import {
    RiBankCardLine,
    RiCalendarEventLine,
    RiErrorWarningLine,
    RiShieldUserLine,
    RiTicketLine,
    RiTimeLine,
} from "@remixicon/react"

import { PricingClient } from "@/components/pricing-client"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getMe } from "@/lib/client-api/auth"
import { listEvents } from "@/lib/client-api/events"
import { listOrgSubscriptions } from "@/lib/client-api/plans"
import { queryKeys } from "@/react-query/query-keys"

const EMPTY_SUBS = {
  subscriptions: [] as never[],
  total_quota: null,
  used_quota: 0,
  unlimited: false,
}

export default function DashboardPage() {
    const { data: user } = useQuery({
        queryKey: queryKeys.me(),
        queryFn: getMe,
    })

    const { data: subs = EMPTY_SUBS } = useQuery({
        queryKey: queryKeys.orgSubscriptions(),
        queryFn: listOrgSubscriptions,
        retry: false,
    })

    const { data: events = [] } = useQuery({
        queryKey: queryKeys.events(),
        queryFn: listEvents,
        retry: false,
    })

    if (!user) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        )
    }

    return (
        <div className="flex flex-col gap-8">
            {/* 1. Hello & Welcome Message */}
            <div className="flex flex-col gap-1.5">
                <h1 className="font-heading text-2xl font-bold tracking-tight sm:text-3xl text-foreground">
                    Hello & Welcome, {user.name}! 👋
                </h1>
                <p className="text-sm text-muted-foreground">
                    Here is your organization&apos;s overview, event statistics, and active subscription details.
                </p>
            </div>

            {/* 2. Stats Cards Grid */}
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                {/* Stat 1: Total Events */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Total Events
                        </CardTitle>
                        <RiCalendarEventLine className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">{events.length}</div>
                        <p className="text-xs text-muted-foreground mt-1">Events created</p>
                    </CardContent>
                </Card>

                {/* Stat 2: Quota Usage */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Event Quota Used
                        </CardTitle>
                        <RiTicketLine className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {subs.unlimited
                                ? "Unlimited"
                                : `${subs.used_quota} / ${subs.total_quota ?? 0}`}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            {subs.unlimited ? "No event creation limits" : "Allocated event slots"}
                        </p>
                    </CardContent>
                </Card>

                {/* Stat 3: Purchased Subscriptions */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Active Plans
                        </CardTitle>
                        <RiBankCardLine className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {subs.subscriptions.length || 1}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            Active & past subscriptions
                        </p>
                    </CardContent>
                </Card>

                {/* Stat 4: Organization Role */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Your Role
                        </CardTitle>
                        <RiShieldUserLine className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground capitalize">
                            {user.roles?.[0]?.replace("_", " ") || "Member"}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1 truncate">
                            {user.organization_name || "Organization"}
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* 3. Current Plan Details Section (Always visible for UI development) */}
            <div className="flex flex-col gap-4">
                <h2 className="font-heading text-lg font-bold text-foreground">
                    Current Plan Details
                </h2>

                {/* Active Plan Card Example */}
                <Card className="border-border/80 bg-card shadow-2xs transition-all">
                    <CardHeader className="flex flex-row items-center justify-between border-b pb-4">
                        <div>
                            <div className="flex items-center gap-2">
                                <CardTitle className="text-lg font-bold text-foreground">
                                    Base Plan
                                </CardTitle>
                                <Badge variant="default" className="rounded-full px-2.5 py-0.5 text-xs font-semibold">
                                    Active Plan
                                </Badge>
                            </div>
                            <CardDescription className="mt-1 text-xs">
                                Active recurring subscription for {user.organization_name || "your organization"}
                            </CardDescription>
                        </div>
                        <Button variant="outline" size="sm" render={<Link href="/dashboard/billing" />}>
                            Manage Plan
                        </Button>
                    </CardHeader>

                    <CardContent className="pt-6">
                        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                    Purchased Date
                                </span>
                                <span className="text-sm font-semibold text-foreground">
                                    Jan 15, 2026
                                </span>
                            </div>

                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                    Expiry Date
                                </span>
                                <span className="text-sm font-semibold text-foreground">
                                    Feb 15, 2026
                                </span>
                            </div>

                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                    Days Remaining
                                </span>
                                <div className="flex items-center gap-1.5">
                                    <RiTimeLine className="size-4 text-emerald-500" />
                                    <span className="text-sm font-bold text-foreground">
                                        12 days left
                                    </span>
                                </div>
                            </div>

                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                    Plan Amount
                                </span>
                                <span className="text-sm font-bold text-foreground">
                                    $49 / month
                                </span>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Expired Plan Card Example (Ghosted/Dull Style with Highlighted Renew Button) */}
                <Card className="opacity-75 bg-muted/40 border-destructive/40 text-muted-foreground grayscale-20 transition-all">
                    <CardHeader className="flex flex-row items-center justify-between border-b border-border/50 pb-4">
                        <div>
                            <div className="flex items-center gap-2">
                                <CardTitle className="text-lg font-bold text-muted-foreground">
                                    Flash Plan
                                </CardTitle>
                                <Badge variant="destructive" className="rounded-full px-2.5 py-0.5 text-xs font-semibold">
                                    Expired / Past Due
                                </Badge>
                            </div>
                            <CardDescription className="mt-1 text-xs text-muted-foreground/80">
                                Expired subscription pass — renewal required to create events
                            </CardDescription>
                        </div>

                        {/* Highlighted Animated Renew Button */}
                        <Button
                            variant="destructive"
                            size="sm"
                            className="rounded-full px-5 py-2 font-bold shadow-md ring-2 ring-destructive/30 animate-pulse"
                            render={<Link href="/dashboard/billing" />}
                        >
                            <RiErrorWarningLine className="mr-1.5 size-4" />
                            Renew Plan Now
                        </Button>
                    </CardHeader>

                    <CardContent className="pt-6">
                        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground/80 uppercase tracking-wider">
                                    Purchased Date
                                </span>
                                <span className="text-sm font-semibold text-muted-foreground">
                                    Dec 01, 2025
                                </span>
                            </div>

                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground/80 uppercase tracking-wider">
                                    Expiry Date
                                </span>
                                <span className="text-sm font-semibold text-muted-foreground">
                                    Dec 31, 2025
                                </span>
                            </div>

                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground/80 uppercase tracking-wider">
                                    Days Remaining
                                </span>
                                <div className="flex items-center gap-1.5">
                                    <RiTimeLine className="size-4 text-destructive" />
                                    <span className="text-sm font-bold text-destructive">
                                        Expired (0 days left)
                                    </span>
                                </div>
                            </div>

                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground/80 uppercase tracking-wider">
                                    Plan Amount
                                </span>
                                <span className="text-sm font-bold text-muted-foreground">
                                    $29 (One-time)
                                </span>
                            </div>
                        </div>

                        <div className="mt-4 rounded-lg bg-destructive/10 p-3 text-xs text-destructive border border-destructive/20 flex items-center justify-between">
                            <span>
                                <strong>Warning:</strong> Your plan has expired. Event creation is restricted until renewed.
                            </span>
                            <Link href="/dashboard/billing" className="font-semibold underline ml-2">
                                Renew immediately →
                            </Link>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* 4. Rerender Pricing / Plan Component */}
            <div className="rounded-2xl border border-border/80 bg-card/50 shadow-2xs overflow-hidden">
                <div className="border-b px-6 py-4 bg-muted/30 flex items-center justify-between">
                    <div>
                        <h2 className="font-heading text-lg font-bold text-foreground">
                            Available Subscription Plans
                        </h2>
                        <p className="text-xs text-muted-foreground">
                            Compare plans, customize event quotas, or select a new plan for your organization.
                        </p>
                    </div>
                    <Button variant="outline" size="sm" render={<Link href="/dashboard/billing" />}>
                        Go to Billing
                    </Button>
                </div>

                <PricingClient />
            </div>
        </div>
    )
}