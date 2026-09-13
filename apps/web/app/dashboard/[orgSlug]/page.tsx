"use client"

import Link from "next/link"
import { useParams } from "next/navigation"
import {
    RiAlertLine,
    RiBankCardLine,
    RiCalendarEventLine,
    RiShieldUserLine,
    RiTicketLine,
    RiTimeLine,
} from "@remixicon/react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import { useEventsListQuery } from "@/query-hooks/events.api"
import { useBillingOverviewQuery } from "@/query-hooks/plans.api"
import useOrganization from "@/hooks/use-organization"
import useUser from "@/hooks/use-user"

export default function DashboardPage() {
    const orgSlug = useParams<{ orgSlug?: string }>().orgSlug
    const { isOwner, currentOrg } = useOrganization(orgSlug ?? null)
    const { user } = useUser()

    const { data: billing } = useBillingOverviewQuery(isOwner)
    const { data: events = [] } = useEventsListQuery()

    const creditBalance = billing?.credit_balance ?? 0
    const annualStatus = billing?.annual_fee_status ?? "free"

    return (
        <div className="flex flex-col gap-8">
            {/* Overdue Maintenance Banner (Admin only) */}
            {isOwner && annualStatus === "past_due" && (
                <div className="flex items-center justify-between gap-4 rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive dark:bg-destructive/20">
                    <div className="flex items-center gap-3">
                        <RiAlertLine className="size-5 shrink-0" />
                        <div>
                            <p className="font-semibold">
                                Annual Maintenance Overdue
                            </p>
                            <p className="text-xs text-destructive/80">
                                Your organization retains {events.length} event
                                {events.length === 1 ? "" : "s"}. Annual
                                maintenance is past due and new event creation
                                is paused.
                            </p>
                        </div>
                    </div>
                    <Button
                        size="sm"
                        variant="destructive"
                        render={<Link href={`/dashboard/${orgSlug}/billing`} />}
                    >
                        Renew Now
                    </Button>
                </div>
            )}

            {/* 1. Hello & Welcome Message */}
            <div className="flex flex-col gap-1.5">
                <h1 className="font-heading text-2xl font-bold tracking-tight sm:text-3xl text-foreground">
                    Hello & Welcome, {user?.name}! 👋
                </h1>
                <p className="text-sm text-muted-foreground">
                    {
                        isOwner
                            ? "Here is your organization's overview, event statistics, and billing credits."
                            : "Here is your organization's overview and event statistics."
                    }
                </p>
            </div>

            {/* 2. Stats Cards Grid */}
            <div className={
                isOwner ? "grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4"
                    : "grid grid-cols-1 gap-4 sm:grid-cols-2"
            }>
                {/* Stat 1: Total Events */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Total Events
                        </CardTitle>
                        <RiCalendarEventLine className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {events.length}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            Events created
                        </p>
                    </CardContent>
                </Card>

                {/* Stat 2: Credits Usage (Admin only) */}
                {isOwner && (
                    <Card className="shadow-2xs">
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                Available Credits
                            </CardTitle>
                            <RiTicketLine className="size-4 text-primary" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold text-foreground">
                                {creditBalance}
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                                1 credit = 1 new event
                            </p>
                        </CardContent>
                    </Card>
                )}

                {/* Stat 3: Annual Platform Status (Admin only) */}
                {isOwner && (
                    <Card className="shadow-2xs">
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                Annual Status
                            </CardTitle>
                            <RiBankCardLine className="size-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold text-foreground capitalize">
                                {annualStatus === "free"
                                    ? "Free Tier"
                                    : annualStatus === "active"
                                        ? "Active"
                                        : "Past Due"}
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                                {events.length > 0
                                    ? "Events preserved"
                                    : "Zero events (100% Free)"}
                            </p>
                        </CardContent>
                    </Card>
                )}

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
                            {currentOrg?.role?.replace("_", " ") || "Member"}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1 truncate">
                            {currentOrg?.name || "Organization"}
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* 3. Billing & Credits Summary (Admin only) */}
            {isOwner && (
                <Card className="border-border/80 bg-card shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between border-b pb-4">
                        <div>
                            <div className="flex items-center gap-2">
                                <CardTitle className="text-lg font-bold text-foreground">
                                    Event Credits & Platform Status
                                </CardTitle>
                                <Badge
                                    variant={
                                        annualStatus === "active"
                                            ? "default"
                                            : annualStatus === "past_due"
                                                ? "destructive"
                                                : "secondary"
                                    }
                                    className="rounded-lg px-2.5 py-0.5 text-xs font-semibold uppercase"
                                >
                                    {annualStatus === "free"
                                        ? "Free (0 Events)"
                                        : annualStatus}
                                </Badge>
                            </div>
                            <CardDescription className="mt-1 text-xs">
                                Each event creation costs 1 Event Credit.
                                Accounts with 0 events are 100% free forever.
                            </CardDescription>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button
                                size="sm"
                                render={
                                    <Link
                                        href={`/dashboard/${orgSlug}/billing`}
                                    />
                                }
                            >
                                Buy Credits
                            </Button>
                            <Button
                                variant="outline"
                                size="sm"
                                render={
                                    <Link
                                        href={`/dashboard/${orgSlug}/billing`}
                                    />
                                }
                            >
                                Manage Billing
                            </Button>
                        </div>
                    </CardHeader>

                    <CardContent className="pt-6">
                        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                    Credit Balance
                                </span>
                                <span className="text-lg font-bold text-foreground">
                                    {creditBalance} Available Credit
                                    {creditBalance === 1 ? "" : "s"}
                                </span>
                            </div>

                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                    Annual Renewal
                                </span>
                                <div className="flex items-center gap-1.5">
                                    <RiTimeLine className="size-4 text-emerald-500" />
                                    <span className="text-sm font-medium text-foreground">
                                        {billing?.current_period_end
                                            ? billing.current_period_end
                                            : "N/A (Free tier)"}
                                    </span>
                                </div>
                            </div>

                            <div className="flex flex-col gap-1">
                                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                    Retained Events
                                </span>
                                <span className="text-sm font-medium text-foreground">
                                    {events.length} event
                                    {events.length === 1 ? "" : "s"} in history
                                </span>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    )
}
