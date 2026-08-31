"use client";

import { useMemo } from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { toast } from "sonner";
import {
    RiBankCardLine,
    RiDownloadLine,
    RiErrorWarningLine,
    RiTicket2Line,
} from "@remixicon/react";

import { RenewButton } from "@/app/dashboard/billing/renew-button";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DataTable } from "@/components/data-table";
import { SubscribeDialog } from "@/components/subscribe-dialog";
import { useOrgSettingsQuery } from "@/query-hooks/organizations.api";
import { useListBillingQuery, useListPlansQuery } from "@/query-hooks/plans.api";
import type { Billing } from "@/schema/plans.types";

type TransactionItem = {
    id: string;
    invoiceNo: string;
    planCode: string;
    kind: Billing["kind"];
    billingCycle: string;
    billingNumber: number;
    status: Billing["status"];
    amount: string;
    periodStart: string;
    periodEnd: string;
    paidAt?: string | null;
};

function formatDate(value: string) {
    if (!value) return "N/A";
    return new Date(value).toLocaleDateString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
    });
}

export default function BillingPage() {
    const plansQuery = useListPlansQuery();
    const subsQuery = useListBillingQuery();
    const orgQuery = useOrgSettingsQuery();

    const plans = plansQuery.data ?? [];
    const subs =
        subsQuery.data ?? { billings: [], credits: [], available_by_type: {} };
    const organizationName = orgQuery.data?.name || "IdentityCard Org";

    const pastDue = subs.billings.filter((b) => b.status === "pending");

    // Generate Transaction records from organization billings
    const transactions = useMemo<TransactionItem[]>(() => {
        return subs.billings.map((b, index) => {
            const shortId = b.id.replace(/-/g, "").slice(0, 4).toUpperCase();
            const invoiceNo = `INV-2026-${shortId}${100 + index}`;
            const amountStr = `$${(b.amount / 100).toFixed(2)} ${b.currency.toUpperCase()}`;

            return {
                id: b.id,
                invoiceNo,
                planCode: b.plan_code,
                kind: b.kind,
                billingCycle: b.billing_cycle,
                billingNumber: b.billing_number,
                status: b.status,
                amount: amountStr,
                periodStart: b.period_start,
                periodEnd: b.period_end,
                paidAt: b.paid_at,
            };
        });
    }, [subs.billings]);

    // Download Invoice File
    function handleDownloadInvoice(tx: TransactionItem) {
        const content = `=====================================================
OFFICIAL INVOICE & PAYMENT RECEIPT
IdentityCard Event Management Platform
=====================================================

Invoice Number:   ${tx.invoiceNo}
Transaction ID:   TXN-${tx.id}
Date Issued:      ${formatDate(tx.periodStart)}
Organization:     ${organizationName}
Status:           ${tx.status.toUpperCase()}

-----------------------------------------------------
PURCHASE DETAILS
-----------------------------------------------------
Plan Code:        ${tx.planCode.toUpperCase()} (${tx.kind})
Billing Cycle:    ${tx.billingCycle}
Cycle Number:     #${tx.billingNumber}
Period Starts:    ${formatDate(tx.periodStart)}
Period Ends:      ${formatDate(tx.periodEnd)}

-----------------------------------------------------
AMOUNT PAID:      ${tx.amount}
Payment Method:   Payment Gateway
-----------------------------------------------------

Thank you for choosing IdentityCard!
For support inquiries, contact billing@identitycard.io
=====================================================`;

        const blob = new Blob([content], { type: "text/plain;charset=utf-8" });
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = `Invoice-${tx.invoiceNo}.txt`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);

        toast.success(`Invoice ${tx.invoiceNo} downloaded successfully!`);
    }

    // Columns definition for Transactions DataTable
    const columns = useMemo<ColumnDef<TransactionItem>[]>(
        () => [
            {
                accessorKey: "invoiceNo",
                header: "Invoice / Transaction",
                cell: ({ row }) => {
                    const tx = row.original;
                    return (
                        <div className="flex items-center gap-3">
                            <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                                <RiBankCardLine className="size-5" />
                            </div>
                            <div className="flex flex-col">
                                <span className="font-semibold text-foreground capitalize">
                                    {tx.planCode} Plan
                                </span>
                                <span className="text-[11px] font-mono text-muted-foreground">
                                    {tx.invoiceNo} • Cycle #{tx.billingNumber}
                                </span>
                            </div>
                        </div>
                    );
                },
            },
            {
                accessorKey: "billingCycle",
                header: "Billing Cycle",
                cell: ({ row }) => {
                    return (
                        <Badge variant="outline" className="text-xs font-medium capitalize">
                            {row.original.billingCycle}
                        </Badge>
                    );
                },
            },
            {
                accessorKey: "amount",
                header: "Amount Paid",
                cell: ({ row }) => {
                    return (
                        <span className="font-mono text-sm font-bold text-foreground">
                            {row.original.amount}
                        </span>
                    );
                },
            },
            {
                accessorKey: "status",
                header: "Status",
                cell: ({ row }) => {
                    const status = row.original.status;
                    return (
                        <Badge
                            variant="outline"
                            className={`rounded-lg px-3 py-0.5 text-xs font-semibold capitalize ${status === "active"
                                ? "bg-emerald-100 text-emerald-900 border-emerald-300 dark:bg-emerald-950/80 dark:text-emerald-300 dark:border-emerald-700"
                                : status === "pending"
                                    ? "bg-amber-100 text-amber-900 border-amber-300 dark:bg-amber-950/80 dark:text-amber-300 dark:border-amber-700"
                                    : "bg-muted text-muted-foreground border-border"
                                }`}
                        >
                            {status}
                        </Badge>
                    );
                },
            },
            {
                accessorKey: "periodStart",
                header: "Date & Period",
                cell: ({ row }) => {
                    const tx = row.original;
                    return (
                        <div className="flex flex-col text-xs text-muted-foreground">
                            <span>Start: {formatDate(tx.periodStart)}</span>
                            <span className="text-[11px] opacity-80">
                                End: {formatDate(tx.periodEnd)}
                            </span>
                        </div>
                    );
                },
            },
            {
                id: "actions",
                header: () => <div className="text-right">Invoice</div>,
                cell: ({ row }) => {
                    const tx = row.original;
                    return (
                        <div className="flex items-center justify-end text-right">
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleDownloadInvoice(tx)}
                                className="h-8 gap-1.5 px-3 text-xs font-medium border-primary/40 text-primary hover:bg-primary/5"
                            >
                                <RiDownloadLine className="size-3.5" />
                                Download Invoice
                            </Button>
                        </div>
                    );
                },
            },
        ],
        [organizationName]
    );

    const totalCredits = subs.credits.length;
    const availableCredits = Object.values(subs.available_by_type).reduce((acc: number, v: unknown) => acc + (typeof v === "number" ? v : 0), 0);

    return (
        <div className="flex flex-col gap-8">
            {/* Header Toolbar */}
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <h1 className="font-heading text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
                        Billing & Subscriptions
                    </h1>
                    <p className="mt-1 text-sm text-muted-foreground">
                        Manage organization subscription plans, event credits, and billing records.
                    </p>
                </div>

                <SubscribeDialog
                    plans={plans}
                    title={subs.billings.length ? "Buy / Upgrade Plan" : "Choose a Plan"}
                    trigger={
                        <Button className="font-semibold">
                            <RiTicket2Line className="mr-1.5 size-4" />
                            {subs.billings.length ? "Buy / Upgrade Plan" : "Choose a Plan"}
                        </Button>
                    }
                />
            </div>

            {/* Overdue Warning Banners */}
            {pastDue.map((b) => (
                <div
                    key={b.id}
                    className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-lg border border-destructive/40 bg-destructive/5 dark:bg-destructive/10 p-4 shadow-2xs"
                >
                    <div className="flex items-start gap-3">
                        <RiErrorWarningLine className="mt-0.5 size-5 shrink-0 text-destructive" />
                        <div>
                            <p className="text-sm font-semibold text-destructive">
                                Renewal pending for {b.plan_code} plan
                            </p>
                            <p className="text-xs text-muted-foreground mt-0.5">
                                Your billing cycle #{b.billing_number} is pending payment. Please renew to continue creating events.
                            </p>
                        </div>
                    </div>
                    <RenewButton lineageRootId={b.lineage_root_id} />
                </div>
            ))}

            {/* Overview Stats Cards */}
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <Card className="shadow-2xs">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Event Credits
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {availableCredits} / {totalCredits} Available
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            Usable event credit balance
                        </p>
                    </CardContent>
                </Card>

                <Card className="shadow-2xs">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Active Billings
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {subs.billings.filter((b) => b.status === "active").length} Active
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            {subs.billings.length} total billing cycle(s)
                        </p>
                    </CardContent>
                </Card>

                <Card className="shadow-2xs">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Total Invoiced Transactions
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {transactions.length} Invoice(s)
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            All transactions recorded with downloadable invoices
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* Transactions & Invoices DataTable */}
            <div className="flex flex-col gap-4">
                <div>
                    <h2 className="text-lg font-bold tracking-tight text-foreground">
                        Transactions & Invoices
                    </h2>
                    <p className="text-xs text-muted-foreground">
                        Complete transaction history and instant invoice downloads.
                    </p>
                </div>

                <DataTable
                    columns={columns}
                    data={transactions}
                    searchPlaceholder="Search invoices or plans..."
                    emptyMessage="No billing transactions found."
                />
            </div>
        </div>
    );
}