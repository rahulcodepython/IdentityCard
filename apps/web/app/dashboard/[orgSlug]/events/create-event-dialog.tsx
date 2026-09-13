"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { RiAlertLine, RiTicket2Line } from "@remixicon/react";

import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { useBillingOverviewQuery } from "@/query-hooks/plans.api";
import useOrganization from "@/hooks/use-organization";

import { CreateEventForm } from "./create-event-form";

export function CreateEventDialog() {
    const [open, setOpen] = useState(false);
    const params = useParams<{ orgSlug: string }>();
    const orgSlug = params?.orgSlug || "";
    const { isOwner } = useOrganization(orgSlug);
    const { data: billing } = useBillingOverviewQuery(isOwner);

    const isPastDue = isOwner && billing?.annual_fee_status === "past_due";
    const hasZeroCredits = isOwner && billing !== undefined && billing.credit_balance <= 0;

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button>New event</Button>} />
            <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
                <DialogHeader>
                    <DialogTitle>Create event</DialogTitle>
                    <DialogDescription>
                        Set up the schedule — each event creation consumes 1 Event Credit.
                    </DialogDescription>
                </DialogHeader>

                {isPastDue ? (
                    <div className="flex flex-col gap-4 py-2">
                        <div className="flex items-center gap-3 rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive dark:bg-destructive/20">
                            <RiAlertLine className="size-5 shrink-0" />
                            <div className="flex-1">
                                <p className="font-semibold">Annual Maintenance Overdue</p>
                                <p className="text-xs text-destructive/80 mt-0.5">
                                    Your organization has an overdue annual maintenance fee. Event creation is paused until the annual renewal is completed.
                                </p>
                            </div>
                        </div>
                        <Button
                            className="w-full"
                            variant="destructive"
                            render={<Link href={`/dashboard/${orgSlug}/billing`} />}
                            onClick={() => setOpen(false)}
                        >
                            Renew Annual Maintenance
                        </Button>
                    </div>
                ) : hasZeroCredits ? (
                    <div className="flex flex-col gap-4 py-2">
                        <div className="flex items-center gap-3 rounded-lg border border-amber-500/30 bg-amber-500/10 p-4 text-sm text-amber-900 dark:text-amber-200 dark:bg-amber-500/20">
                            <RiTicket2Line className="size-5 shrink-0 text-amber-600 dark:text-amber-400" />
                            <div className="flex-1">
                                <p className="font-semibold">0 Event Credits Available</p>
                                <p className="text-xs text-amber-800 dark:text-amber-300 mt-0.5">
                                    Each event creation requires 1 Event Credit. Please purchase credits to continue.
                                </p>
                            </div>
                        </div>
                        <Button
                            className="w-full"
                            render={<Link href={`/dashboard/${orgSlug}/billing`} />}
                            onClick={() => setOpen(false)}
                        >
                            Purchase Event Credits
                        </Button>
                    </div>
                ) : (
                    <CreateEventForm />
                )}
            </DialogContent>
        </Dialog>
    );
}
