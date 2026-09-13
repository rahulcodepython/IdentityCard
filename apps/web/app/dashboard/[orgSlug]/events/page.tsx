"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { RiAlertLine, RiCoinsLine } from "@remixicon/react";

import { useEventsListQuery } from "@/query-hooks/events.api";
import { useBillingOverviewQuery } from "@/query-hooks/plans.api";
import useOrganization from "@/hooks/use-organization";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import { CreateEventDialog } from "./create-event-dialog";
import { EventsTable } from "./events-table";

export default function EventsPage() {
    const params = useParams<{ orgSlug: string }>();
    const orgSlug = params?.orgSlug || "";
    const { isOwner } = useOrganization(orgSlug);
    const { data: events = [] } = useEventsListQuery();
    const { data: billing } = useBillingOverviewQuery(isOwner);

    const isPastDue = isOwner && billing?.annual_fee_status === "past_due";
    const hasZeroCredits = isOwner && billing !== undefined && billing.credit_balance <= 0;

    return (
        <div className="flex flex-col gap-6">
            {/* Maintenance Overdue Alert */}
            {isPastDue && (
                <div className="flex items-center justify-between gap-4 rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive dark:bg-destructive/20">
                    <div className="flex items-center gap-3">
                        <RiAlertLine className="size-5 shrink-0" />
                        <div>
                            <p className="font-semibold">Annual Maintenance Overdue</p>
                            <p className="text-xs text-destructive/80">
                                Event creation is paused because your annual platform fee has lapsed. Please renew to resume creating events.
                            </p>
                        </div>
                    </div>
                    <Button size="sm" variant="destructive" render={<Link href={`/dashboard/${orgSlug}/billing`} />}>
                        Renew Maintenance
                    </Button>
                </div>
            )}

            {/* Zero Credits Alert */}
            {!isPastDue && hasZeroCredits && (
                <div className="flex items-center justify-between gap-4 rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-sm text-amber-900 dark:text-amber-200 dark:bg-amber-500/20">
                    <div className="flex items-center gap-3">
                        <RiCoinsLine className="size-5 shrink-0 text-amber-600 dark:text-amber-400" />
                        <div>
                            <p className="font-semibold">No Event Credits Remaining</p>
                            <p className="text-xs text-amber-800 dark:text-amber-300">
                                You have 0 available credits. Each event requires 1 credit. Purchase credits anytime to launch your next event.
                            </p>
                        </div>
                    </div>
                    <Button size="sm" render={<Link href={`/dashboard/${orgSlug}/billing`} />}>
                        Buy Event Credits
                    </Button>
                </div>
            )}

            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <h1 className="font-heading text-xl font-medium">Events</h1>
                    {isOwner && billing && (
                        <Link href={`/dashboard/${orgSlug}/billing`}>
                            <Badge
                                variant="outline"
                                className="gap-1.5 cursor-pointer hover:bg-muted font-normal text-xs"
                            >
                                <RiCoinsLine className="size-3.5 text-primary" />
                                <span>{billing.credit_balance} {billing.credit_balance === 1 ? "credit" : "credits"}</span>
                            </Badge>
                        </Link>
                    )}
                </div>
                <CreateEventDialog />
            </div>

            <EventsTable data={events} />
        </div>
    );
}