"use client";

import Link from "next/link";
import {
    RiBankCardLine,
    RiCalendarEventLine,
    RiErrorWarningLine,
    RiShieldUserLine,
    RiTicketLine,
    RiTimeLine,
} from "@remixicon/react";

import { PricingClient } from "@/components/pricing-client";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { authClient } from "@/lib/auth-client";
import { useEventsListQuery } from "@/query-hooks/events.api";
import { useListBillingQuery } from "@/query-hooks/plans.api";
import { useSessionStore } from "@/store/session.store";

const EMPTY_SUBS = {
    billings: [] as never[],
    credits: [] as never[],
    available_by_type: {},
};

export default function DashboardPage() {
    const user = useSessionStore((s) => s.user);
    const role = useSessionStore((s) => s.role);
    const { data: organization } = authClient.useActiveOrganization();

    const { data: subs = EMPTY_SUBS } = useListBillingQuery();
    const { data: events = [] } = useEventsListQuery();

    if (!user) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    const totalCredits = subs.credits.length;
    const availableCredits = Object.values(subs.available_by_type).reduce((acc: number, v: unknown) => acc + (typeof v === "number" ? v : 0), 0);

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

                {/* Stat 2: Credits Usage */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Available Credits
                        </CardTitle>
                        <RiTicketLine className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {availableCredits} / {totalCredits}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            Available event slots
                        </p>
                    </CardContent>
                </Card>

                {/* Stat 3: Purchased Subscriptions */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Active Billings
                        </CardTitle>
                        <RiBankCardLine className="size-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {subs.billings.length || 0}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            Active & past billing cycles
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
                            {role?.replace("_", " ") || "Member"}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1 truncate">
                            {organization?.name || "Organization"}
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* 3. Current Plan Details Section */}
            <div className="flex flex-col gap-4">
                <h2 className="font-heading text-lg font-bold text-foreground">
                    Current Plan Details
                </h2>

                {subs.billings.length > 0 ? (
                    subs.billings.map((b) => (
                        <Card key={b.id} className="border-border/80 bg-card shadow-2xs transition-all">
                            <CardHeader className="flex flex-row items-center justify-between border-b pb-4">
                                <div>
                                    <div className="flex items-center gap-2">
                                        <CardTitle className="text-lg font-bold text-foreground capitalize">
                                            {b.plan_code} Plan
                                        </CardTitle>
                                        <Badge
                                            variant={b.status === "active" ? "default" : "destructive"}
                                            className="rounded-lg px-2.5 py-0.5 text-xs font-semibold uppercase"
                                        >
                                            {b.status}
                                        </Badge>
                                    </div>
                                    <CardDescription className="mt-1 text-xs">
                                        Billing cycle #{b.billing_number} • {b.kind} ({b.billing_cycle})
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
                                            Period Start
                                        </span>
                                        <span className="text-sm font-semibold text-foreground">
                                            {b.period_start}
                                        </span>
                                    </div>

                                    <div className="flex flex-col gap-1">
                                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                            Period End
                                        </span>
                                        <span className="text-sm font-semibold text-foreground">
                                            {b.period_end}
                                        </span>
                                    </div>

                                    <div className="flex flex-col gap-1">
                                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                            Status
                                        </span>
                                        <div className="flex items-center gap-1.5">
                                            <RiTimeLine className="size-4 text-emerald-500" />
                                            <span className="text-sm font-bold text-foreground capitalize">
                                                {b.status}
                                            </span>
                                        </div>
                                    </div>

                                    <div className="flex flex-col gap-1">
                                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                            Amount
                                        </span>
                                        <span className="text-sm font-bold text-foreground">
                                            {(b.amount / 100).toFixed(2)} {b.currency}
                                        </span>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    ))
                ) : (
                    <Card className="border-border/80 bg-card shadow-2xs">
                        <CardContent className="py-8 text-center text-sm text-muted-foreground">
                            No active plan subscriptions found. Select a plan below to get started.
                        </CardContent>
                    </Card>
                )}
            </div>

            {/* 4. Rerender Pricing / Plan Component */}
            <div className="rounded-lg border border-border/80 bg-card/50 shadow-2xs overflow-hidden">
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
    );
}