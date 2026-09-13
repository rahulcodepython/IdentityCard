"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { type ColumnDef } from "@tanstack/react-table";
import { useSessionStore } from "@/store/session.store";
import useOrganization from "@/hooks/use-organization";
import {
    RiAddLine,
    RiAlertLine,
    RiBankCardLine,
    RiCalendarCheckLine,
    RiDownloadLine,
    RiInformationLine,
    RiLock2Line,
    RiSubtractLine,
    RiTicket2Line,
} from "@remixicon/react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { DataTable } from "@/components/table/data-table";
import { useOrgSettingsQuery } from "@/query-hooks/organizations.api";
import {
    useBillingOverviewQuery,
    usePurchaseCreditsMutation,
    useRenewAnnualMutation,
} from "@/query-hooks/plans.api";
import type { BillingTransaction } from "@/schema/plans.types";

function formatDate(value?: string | null) {
    if (!value) return "N/A";
    return new Date(value).toLocaleDateString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
    });
}

function formatCurrency(amount: number, currency = "INR") {
    const symbol = currency === "INR" ? "₹" : "$";
    return `${symbol}${(amount / 100).toLocaleString("en-IN", { minimumFractionDigits: 2 })}`;
}

export default function BillingPage() {
    const params = useParams<{ orgSlug: string }>();
    const orgSlug = params?.orgSlug || "";
    const { isOwner } = useOrganization(orgSlug);

    const orgQuery = useOrgSettingsQuery(undefined, isOwner);
    const billingQuery = useBillingOverviewQuery(isOwner);
    const purchaseMutation = usePurchaseCreditsMutation();
    const renewMutation = useRenewAnnualMutation();

    const [buyCreditsOpen, setBuyCreditsOpen] = useState(false);
    const [quantity, setQuantity] = useState(1);

    const billing = billingQuery.data ?? {
        credit_balance: 0,
        annual_fee_status: "free",
        current_period_start: null,
        current_period_end: null,
        event_count: 0,
        pricing: {
            credit_unit_price: 149900,
            currency: "INR",
            annual_renewal_amount: 299900,
        },
        transactions: [],
    };

    const organizationName = orgQuery.data?.name || "IdentityCard Org";
    const unitPrice = billing.pricing.credit_unit_price;
    const totalAmount = quantity * unitPrice;

    async function handleBuyCredits() {
        await purchaseMutation.execute({ quantity });
        setBuyCreditsOpen(false);
        setQuantity(1);
    }

    async function handleRenew() {
        await renewMutation.execute();
    }

    function handleDownloadReceipt(tx: BillingTransaction) {
        const content = `=====================================================
OFFICIAL PAYMENT RECEIPT
IdentityCard Event Management Platform
=====================================================

Receipt ID:       ${tx.id}
Date Issued:      ${formatDate(tx.created_at)}
Organization:     ${organizationName}
Transaction Type: ${tx.type.replace("_", " ").toUpperCase()}

-----------------------------------------------------
DETAILS
-----------------------------------------------------
Credits Delta:    ${tx.credits_delta > 0 ? `+${tx.credits_delta}` : tx.credits_delta}
Amount Paid:      ${formatCurrency(tx.amount, tx.currency)}
Description:      ${tx.description || "N/A"}
Status:           COMPLETED
=====================================================
`;
        const blob = new Blob([content], { type: "text/plain" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `receipt-${tx.id.slice(0, 8)}.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    const columns: ColumnDef<BillingTransaction>[] = [
        {
            accessorKey: "created_at",
            header: "Date",
            cell: ({ row }) => (
                <span className="text-xs font-medium text-foreground">
                    {formatDate(row.original.created_at)}
                </span>
            ),
        },
        {
            accessorKey: "type",
            header: "Type",
            cell: ({ row }) => {
                const type = row.original.type;
                if (type === "credit_purchase") {
                    return (
                        <Badge variant="default" className="bg-emerald-600/10 text-emerald-600 border-emerald-500/20">
                            Credit Purchase
                        </Badge>
                    );
                }
                if (type === "credit_consumed") {
                    return (
                        <Badge variant="secondary" className="bg-muted text-muted-foreground">
                            Credit Consumed
                        </Badge>
                    );
                }
                return (
                    <Badge variant="outline" className="border-primary/20 text-primary">
                        Annual Renewal
                    </Badge>
                );
            },
        },
        {
            accessorKey: "credits_delta",
            header: "Credits",
            cell: ({ row }) => {
                const delta = row.original.credits_delta;
                if (delta > 0) {
                    return <span className="text-xs font-bold text-emerald-600">+{delta}</span>;
                }
                if (delta < 0) {
                    return <span className="text-xs font-bold text-amber-600">{delta}</span>;
                }
                return <span className="text-xs text-muted-foreground">-</span>;
            },
        },
        {
            accessorKey: "amount",
            header: "Amount",
            cell: ({ row }) => (
                <span className="text-xs font-semibold text-foreground">
                    {formatCurrency(row.original.amount, row.original.currency)}
                </span>
            ),
        },
        {
            accessorKey: "description",
            header: "Description",
            cell: ({ row }) => (
                <span className="text-xs text-muted-foreground">
                    {row.original.description || "-"}
                </span>
            ),
        },
        {
            id: "actions",
            header: "Receipt",
            cell: ({ row }) => (
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 px-2 text-xs"
                    onClick={() => handleDownloadReceipt(row.original)}
                >
                    <RiDownloadLine className="size-3.5 mr-1" />
                    Receipt
                </Button>
            ),
        },
    ];

    if (!isOwner) {
        return (
            <div className="flex min-h-[60vh] flex-col items-center justify-center text-center">
                <div className="flex size-14 items-center justify-center rounded-2xl bg-destructive/10 text-destructive mb-4">
                    <RiLock2Line className="size-7" />
                </div>
                <h1 className="font-heading text-2xl font-bold tracking-tight text-foreground">
                    Access Restricted
                </h1>
                <p className="mt-2 max-w-md text-sm text-muted-foreground">
                    You do not have permission to access the billing and credits portal. Only organization administrators can manage billing and purchases.
                </p>
                <div className="mt-6">
                    <Button render={<Link href={`/dashboard/${orgSlug}`} />}>
                        Return to Dashboard
                    </Button>
                </div>
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-6">
            {/* 1. Header */}
            <div>
                <h1 className="font-heading text-2xl font-bold tracking-tight text-foreground">
                    Billing & Event Credits
                </h1>
                <p className="text-sm text-muted-foreground">
                    Manage event creation credits, monitor annual platform status, and review receipts.
                </p>
            </div>

            {/* Overdue Alert Banner */}
            {billing.annual_fee_status === "past_due" && (
                <div className="flex items-center justify-between gap-4 rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive dark:bg-destructive/20">
                    <div className="flex items-center gap-3">
                        <RiAlertLine className="size-5 shrink-0" />
                        <div>
                            <p className="font-semibold">Annual Maintenance Overdue</p>
                            <p className="text-xs text-destructive/80 mt-0.5">
                                Your organization retains {billing.event_count} event{billing.event_count === 1 ? "" : "s"}. Annual maintenance is past due and new event creation is paused. Please renew to restore full privileges.
                            </p>
                        </div>
                    </div>
                    <Button
                        size="sm"
                        variant="destructive"
                        onClick={handleRenew}
                        disabled={renewMutation.isPending}
                    >
                        {renewMutation.isPending ? "Renewing…" : `Pay Renewal (${formatCurrency(billing.pricing.annual_renewal_amount)})`}
                    </Button>
                </div>
            )}

            {/* 2. Top Summary Cards */}
            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                {/* Card 1: Event Credits */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                            Event Credits
                        </CardTitle>
                        <RiTicket2Line className="size-5 text-primary" />
                    </CardHeader>
                    <CardContent className="flex flex-col gap-3">
                        <div className="flex items-baseline gap-2">
                            <span className="text-3xl font-bold text-foreground">
                                {billing.credit_balance}
                            </span>
                            <span className="text-xs text-muted-foreground">
                                credit{billing.credit_balance === 1 ? "" : "s"} available
                            </span>
                        </div>
                        <p className="text-xs text-muted-foreground">
                            Each new event creation consumes exactly 1 Event Credit.
                        </p>
                        <Button
                            size="sm"
                            className="mt-1 w-full"
                            onClick={() => setBuyCreditsOpen(true)}
                        >
                            <RiAddLine className="size-4 mr-1" />
                            Buy Event Credits
                        </Button>
                    </CardContent>
                </Card>

                {/* Card 2: Annual Platform Status */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                            Annual Platform Status
                        </CardTitle>
                        <RiCalendarCheckLine className="size-5 text-emerald-500" />
                    </CardHeader>
                    <CardContent className="flex flex-col gap-3">
                        <div className="flex items-center gap-2">
                            <Badge
                                variant={billing.annual_fee_status === "active" ? "default" : "secondary"}
                                className="capitalize"
                            >
                                {billing.annual_fee_status === "free" ? "Free Tier (0 Events)" : billing.annual_fee_status}
                            </Badge>
                            <span className="text-xs text-muted-foreground">
                                {billing.event_count} event{billing.event_count === 1 ? "" : "s"} retained
                            </span>
                        </div>
                        <div className="text-xs text-muted-foreground">
                            {billing.current_period_end ? (
                                <span>Next Renewal: <strong>{formatDate(billing.current_period_end)}</strong></span>
                            ) : (
                                <span>Zero-event accounts are 100% free with no recurring charges.</span>
                            )}
                        </div>
                        {billing.annual_fee_status === "past_due" && (
                            <Button
                                size="sm"
                                variant="destructive"
                                className="w-full"
                                onClick={handleRenew}
                                disabled={renewMutation.isPending}
                            >
                                {renewMutation.isPending ? "Renewing…" : `Renew Annual Plan (${formatCurrency(billing.pricing.annual_renewal_amount)})`}
                            </Button>
                        )}
                    </CardContent>
                </Card>

                {/* Card 3: Transparent Pricing */}
                <Card className="shadow-2xs">
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                            Transparent Pricing
                        </CardTitle>
                        <RiBankCardLine className="size-5 text-muted-foreground" />
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-xs">
                        <div className="flex justify-between border-b pb-1.5">
                            <span className="text-muted-foreground">Per Event Credit:</span>
                            <span className="font-semibold text-foreground">{formatCurrency(billing.pricing.credit_unit_price)}</span>
                        </div>
                        <div className="flex justify-between border-b pb-1.5">
                            <span className="text-muted-foreground">Zero Events:</span>
                            <span className="font-semibold text-emerald-600">₹0 (Always Free)</span>
                        </div>
                        <div className="flex justify-between pb-1.5">
                            <span className="text-muted-foreground">Annual Retention:</span>
                            <span className="font-semibold text-foreground">{formatCurrency(billing.pricing.annual_renewal_amount)}/year</span>
                        </div>
                        <div className="flex items-center gap-1 text-[11px] text-muted-foreground mt-1">
                            <RiInformationLine className="size-3.5 text-primary shrink-0" />
                            <span>1 free welcome credit is granted upon signup.</span>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* 3. Transaction History Table */}
            <div className="flex flex-col gap-3">
                <div>
                    <h2 className="font-heading text-lg font-bold text-foreground">
                        Billing & Credit History
                    </h2>
                    <p className="text-xs text-muted-foreground">
                        Audit ledger of credit purchases, deductions, and annual subscription transactions.
                    </p>
                </div>

                <DataTable
                    columns={columns}
                    data={billing.transactions}
                    searchPlaceholder="Filter transactions…"
                />
            </div>

            {/* 4. Buy Credits Dialog */}
            <Dialog open={buyCreditsOpen} onOpenChange={setBuyCreditsOpen}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>Purchase Event Credits</DialogTitle>
                        <DialogDescription>
                            Each event creation costs 1 credit. Enter the number of credits you want to purchase.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="flex flex-col gap-4 py-4">
                        <div className="flex items-center justify-between">
                            <span className="text-sm font-medium text-foreground">Credit Quantity</span>
                            <div className="flex items-center gap-2">
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="icon"
                                    className="size-8"
                                    onClick={() => setQuantity((q) => Math.max(1, q - 1))}
                                    disabled={quantity <= 1}
                                >
                                    <RiSubtractLine className="size-4" />
                                </Button>
                                <Input
                                    type="number"
                                    min={1}
                                    max={500}
                                    value={quantity}
                                    onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
                                    className="w-16 text-center font-bold"
                                />
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="icon"
                                    className="size-8"
                                    onClick={() => setQuantity((q) => q + 1)}
                                >
                                    <RiAddLine className="size-4" />
                                </Button>
                            </div>
                        </div>

                        <div className="rounded-lg bg-muted/50 p-3 flex flex-col gap-1.5 text-xs">
                            <div className="flex justify-between">
                                <span className="text-muted-foreground">Price per credit:</span>
                                <span>{formatCurrency(unitPrice)}</span>
                            </div>
                            <div className="flex justify-between font-bold text-sm text-foreground pt-1 border-t">
                                <span>Total Amount:</span>
                                <span className="text-primary">{formatCurrency(totalAmount)}</span>
                            </div>
                        </div>
                    </div>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setBuyCreditsOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleBuyCredits} disabled={purchaseMutation.isPending}>
                            {purchaseMutation.isPending ? "Processing…" : `Pay ${formatCurrency(totalAmount)}`}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}