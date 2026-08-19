"use client"

import { useMemo, useState, useTransition } from "react"
import { type ColumnDef } from "@tanstack/react-table"
import { toast } from "sonner"
import {
    RiBankCardLine,
    RiCheckLine,
    RiDownloadLine,
    RiErrorWarningLine,
    RiPulseLine,
    RiTicket2Line,
} from "@remixicon/react"

import { RenewButton } from "@/app/dashboard/billing/renew-button"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DataTable } from "@/components/ui/data-table"
import { BILLING_CYCLE_LABEL } from "@/components/plan-card"
import { SubscribeDialog } from "@/components/subscribe-dialog"
import type { OrgSubscriptions, Plan, Subscription } from "@/lib/validation/plans"

const KIND_LABEL: Record<Subscription["kind"], string> = {
    flash: "Flash Plan",
    base: "Base Plan",
    custom: "Custom Plan",
    unlimited: "Unlimited Plan",
}

const KIND_DEFAULT_AMOUNT: Record<Subscription["kind"], string> = {
    flash: "$29.00 USD",
    base: "$199.00 USD",
    custom: "$349.00 USD",
    unlimited: "$499.00 USD",
}

export type TransactionItem = {
    id: string
    invoiceNo: string
    planName: string
    kind: Subscription["kind"]
    billingCycle: string
    status: Subscription["status"]
    amount: string
    quotaInfo: string
    startedAt: string
    currentPeriodEnd?: string | null
    graceDeadline?: string | null
}

function formatDate(value: string) {
    return new Date(value).toLocaleDateString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
    })
}

export function BillingClient({
    plans,
    subs,
    organizationName = "IdentityCard Org",
}: {
    plans: Plan[]
    subs: OrgSubscriptions
    organizationName?: string
}) {
    const pastDue = subs.subscriptions.filter((s) => s.status === "past_due")

    // Generate Transaction records from organization subscriptions
    const transactions = useMemo<TransactionItem[]>(() => {
        return subs.subscriptions.map((sub, index) => {
            // Find matching plan amount if available
            const matchedPlan = plans.find(
                (p) => p.kind === sub.kind && p.billing_cycle === sub.billing_cycle
            )

            let amountStr = KIND_DEFAULT_AMOUNT[sub.kind]
            if (matchedPlan?.amount) {
                amountStr = `$${(matchedPlan.amount / 100).toFixed(2)} ${matchedPlan.currency.toUpperCase()}`
            }

            const shortId = sub.id.replace(/-/g, "").slice(0, 4).toUpperCase()
            const invoiceNo = `INV-2026-${shortId}${100 + index}`

            return {
                id: sub.id,
                invoiceNo,
                planName: KIND_LABEL[sub.kind] || "Event Subscription",
                kind: sub.kind,
                billingCycle: BILLING_CYCLE_LABEL[sub.billing_cycle] || "Monthly",
                status: sub.status,
                amount: amountStr,
                quotaInfo:
                    sub.event_quota != null ? `${sub.event_quota} event slot(s)` : "Unlimited events",
                startedAt: sub.started_at,
                currentPeriodEnd: sub.current_period_end,
                graceDeadline: sub.grace_deadline,
            }
        })
    }, [subs.subscriptions, plans])

    // Download Invoice File
    function handleDownloadInvoice(tx: TransactionItem) {
        const content = `=====================================================
OFFICIAL INVOICE & PAYMENT RECEIPT
IdentityCard Event Management Platform
=====================================================

Invoice Number:   ${tx.invoiceNo}
Transaction ID:   TXN-${tx.id}
Date Issued:      ${formatDate(tx.startedAt)}
Organization:     ${organizationName}
Status:           ${tx.status.toUpperCase()}

-----------------------------------------------------
PURCHASE DETAILS
-----------------------------------------------------
Item Description: ${tx.planName}
Billing Cycle:    ${tx.billingCycle}
Capacity Quota:   ${tx.quotaInfo}
Period Ends:      ${tx.currentPeriodEnd ? formatDate(tx.currentPeriodEnd) : "N/A"}

-----------------------------------------------------
AMOUNT PAID:      ${tx.amount}
Payment Method:   Credit Card (Ending in ****4242)
-----------------------------------------------------

Thank you for choosing IdentityCard!
For support inquiries, contact billing@identitycard.io
=====================================================`

        const blob = new Blob([content], { type: "text/plain;charset=utf-8" })
        const url = URL.createObjectURL(blob)
        const link = document.createElement("a")
        link.href = url
        link.download = `Invoice-${tx.invoiceNo}.txt`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(url)

        toast.success(`Invoice ${tx.invoiceNo} downloaded successfully!`)
    }

    // Columns definition for Transactions DataTable
    const columns = useMemo<ColumnDef<TransactionItem>[]>(
        () => [
            {
                accessorKey: "invoiceNo",
                header: "Invoice / Transaction",
                cell: ({ row }) => {
                    const tx = row.original
                    return (
                        <div className="flex items-center gap-3">
                            <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                                <RiBankCardLine className="size-5" />
                            </div>
                            <div className="flex flex-col">
                                <span className="font-semibold text-foreground">
                                    {tx.planName}
                                </span>
                                <span className="text-[11px] font-mono text-muted-foreground">
                                    {tx.invoiceNo}
                                </span>
                            </div>
                        </div>
                    )
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
                    )
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
                    )
                },
            },
            {
                accessorKey: "status",
                header: "Status",
                cell: ({ row }) => {
                    const status = row.original.status
                    return (
                        <Badge
                            variant="outline"
                            className={`rounded-full px-3 py-0.5 text-xs font-semibold capitalize ${status === "active"
                                ? "bg-emerald-100 text-emerald-900 border-emerald-300 dark:bg-emerald-950/80 dark:text-emerald-300 dark:border-emerald-700"
                                : status === "past_due"
                                    ? "bg-red-100 text-red-900 border-red-300 dark:bg-red-950/80 dark:text-red-300 dark:border-red-700"
                                    : "bg-muted text-muted-foreground border-border"
                                }`}
                        >
                            {status.replace("_", " ")}
                        </Badge>
                    )
                },
            },
            {
                accessorKey: "startedAt",
                header: "Date & Renewal",
                cell: ({ row }) => {
                    const tx = row.original
                    return (
                        <div className="flex flex-col text-xs text-muted-foreground">
                            <span>Paid: {formatDate(tx.startedAt)}</span>
                            {tx.currentPeriodEnd && (
                                <span className="text-[11px] opacity-80">
                                    {tx.status === "expired" ? "Ended" : "Renews"}: {formatDate(tx.currentPeriodEnd)}
                                </span>
                            )}
                        </div>
                    )
                },
            },
            {
                id: "actions",
                header: () => <div className="text-right">Invoice</div>,
                cell: ({ row }) => {
                    const tx = row.original
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
                    )
                },
            },
        ],
        []
    )

    return (
        <div className="flex flex-col gap-8">
            {/* Header Toolbar */}
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <h1 className="font-heading text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
                        Billing & Subscriptions
                    </h1>
                    <p className="mt-1 text-sm text-muted-foreground">
                        {subs.unlimited
                            ? "Unlimited events tier active for your organization."
                            : subs.total_quota != null
                                ? `${subs.used_quota} of ${subs.total_quota} total event quota used.`
                                : "No active plan — choose a plan to start creating events."}
                    </p>
                </div>

                <SubscribeDialog
                    plans={plans}
                    title={subs.subscriptions.length ? "Buy / Upgrade Plan" : "Choose a Plan"}
                    trigger={
                        <Button className="font-semibold">
                            <RiTicket2Line className="mr-1.5 size-4" />
                            {subs.subscriptions.length ? "Buy / Upgrade Plan" : "Choose a Plan"}
                        </Button>
                    }
                />
            </div>

            {/* Overdue Warning Banners */}
            {pastDue.map((sub) => (
                <div
                    key={sub.id}
                    className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-xl border border-destructive/40 bg-destructive/5 dark:bg-destructive/10 p-4 shadow-2xs"
                >
                    <div className="flex items-start gap-3">
                        <RiErrorWarningLine className="mt-0.5 size-5 shrink-0 text-destructive" />
                        <div>
                            <p className="text-sm font-semibold text-destructive">
                                Payment overdue for {KIND_LABEL[sub.kind]}
                            </p>
                            <p className="text-xs text-muted-foreground mt-0.5">
                                New event creation is suspended. Please renew by{" "}
                                {sub.grace_deadline ? formatDate(sub.grace_deadline) : "the deadline"}{" "}
                                to prevent plan expiration.
                            </p>
                        </div>
                    </div>
                    <RenewButton subscriptionId={sub.id} />
                </div>
            ))}

            {/* Overview Stats Cards */}
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <Card className="shadow-2xs">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Event Quota Usage
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {subs.unlimited ? "Unlimited" : `${subs.used_quota} / ${subs.total_quota ?? 0}`}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            {subs.unlimited
                                ? "No event limits"
                                : `${(subs.total_quota ?? 0) - subs.used_quota} event slot(s) remaining`}
                        </p>
                    </CardContent>
                </Card>

                <Card className="shadow-2xs">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Active Subscriptions
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-foreground">
                            {subs.subscriptions.filter((s) => s.status === "active").length} Active
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            {subs.subscriptions.length} total plan purchase(s)
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
    )
}
