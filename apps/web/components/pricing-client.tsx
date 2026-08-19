"use client"

import { useState } from "react"
import Link from "next/link"
import { Check, Minus, Plus, Zap, Package, Building2, Crown } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { PlanKind } from "@/lib/validation/plans"

type BillingCycle = "yearly" | "monthly"

type PlanConfig = {
    kind: PlanKind
    name: string
    description: string
    monthlyPrice: number
    yearlyPrice: number
    isPerEvent?: boolean
    badge?: string
    Icon: typeof Zap
    features: (eventCount: number) => string[]
}

const PLANS: PlanConfig[] = [
    {
        kind: "flash",
        name: "Flash",
        description: "Ideal for a single one-day meetup or event",
        monthlyPrice: 29,
        yearlyPrice: 29, // Same amount for monthly & yearly as requested!
        badge: "Same Price",
        Icon: Zap,
        features: () => [
            "1 fixed event creation quota",
            "Digital ID Cards (PDF + QR code)",
            "Scanner device check-in & verification",
            "Standard attendee list dashboard",
            "Basic attendance log reporting",
            "Email support",
        ],
    },
    {
        kind: "base",
        name: "Base",
        description: "Ideal for one event you run again and again",
        monthlyPrice: 49,
        yearlyPrice: 39, // Billed yearly
        badge: "Save 20%",
        Icon: Package,
        features: () => [
            "1 active event slot at a time",
            "Multi-day & sub-event capabilities",
            "Unlimited attendee registrations",
            "Automated email ID card dispatch",
            "Live attendance tracking & analytics",
            "Standard email support",
        ],
    },
    {
        kind: "custom",
        name: "Custom",
        description: "Pick how many events you need — adjust counter anytime",
        monthlyPrice: 15, // per event per month
        yearlyPrice: 12, // per event per month (yearly)
        isPerEvent: true,
        badge: "Save 20%",
        Icon: Building2,
        features: (count) => [
            `${count} fixed event slots included`,
            "Interactive event counter & scaling",
            "Digital ID Cards & unlimited scanners",
            "Advanced analytics, trends & CSV exports",
            "Priority email & chat support",
        ],
    },
    {
        kind: "unlimited",
        name: "Unlimited",
        description: "Ideal for schools, offices, and large organizations",
        monthlyPrice: 199,
        yearlyPrice: 159,
        badge: "Save 20%",
        Icon: Crown,
        features: () => [
            "Unlimited event creation",
            "Unlimited attendees & sub-events",
            "Digital ID Cards (PDF + QR)",
            "Dedicated device pairing & live sync",
            "Full analytics, CSV exports & chart downloads",
            "24/7 Priority support & dedicated manager",
        ],
    },
]

export function PricingClient() {
    const [billingCycle, setBillingCycle] = useState<BillingCycle>("monthly")
    const [selectedKind, setSelectedKind] = useState<PlanKind>("flash")
    const [customEventCount, setCustomEventCount] = useState<number>(5)

    const selectedPlan = PLANS.find((p) => p.kind === selectedKind) ?? PLANS[0]

    // Calculate prices based on selection
    function getDisplayPrice(plan: PlanConfig) {
        if (plan.kind === "flash") {
            return { amount: `$${plan.monthlyPrice}`, period: "/month" }
        }
        if (plan.kind === "custom") {
            const unitRate = billingCycle === "yearly" ? plan.yearlyPrice : plan.monthlyPrice
            const total = unitRate * customEventCount
            return { amount: `$${total}`, period: "/month" }
        }
        const rate = billingCycle === "yearly" ? plan.yearlyPrice : plan.monthlyPrice
        return { amount: `$${rate}`, period: "/month" }
    }

    const selectedPriceInfo = getDisplayPrice(selectedPlan)

    return (
        <div id="pricing" className="px-6 py-20 md:py-28">
            <div className="mx-auto max-w-6xl">
                {/* Header Badge & Title */}
                <div className="flex flex-col items-center text-center">
                    <span className="inline-flex items-center rounded-full border border-border/80 bg-muted/60 px-3.5 py-1 text-xs font-medium text-muted-foreground shadow-2xs">
                        Pricing
                    </span>

                    <h2 className="mt-4 font-heading text-3xl font-bold tracking-tight text-foreground sm:text-4xl md:text-5xl">
                        Flexible Pricing for Every Business
                    </h2>

                    <p className="mt-4 max-w-2xl text-base text-muted-foreground sm:text-lg">
                        Discover a range of pricing plans designed to meet diverse customer
                        requirements, from individuals and small teams to large enterprises.
                    </p>

                    {/* Monthly / Yearly Toggle */}
                    <div className="mt-8 inline-flex items-center rounded-full border border-border/80 bg-muted/60 p-1 shadow-2xs">
                        <button
                            type="button"
                            onClick={() => setBillingCycle("yearly")}
                            className={`rounded-full px-6 py-2 text-xs font-semibold transition-all ${billingCycle === "yearly"
                                ? "bg-foreground text-background shadow-xs"
                                : "text-muted-foreground hover:text-foreground"
                                }`}
                        >
                            Yearly
                        </button>
                        <button
                            type="button"
                            onClick={() => setBillingCycle("monthly")}
                            className={`rounded-full px-6 py-2 text-xs font-semibold transition-all ${billingCycle === "monthly"
                                ? "bg-foreground text-background shadow-xs"
                                : "text-muted-foreground hover:text-foreground"
                                }`}
                        >
                            Monthly
                        </button>
                    </div>
                </div>

                {/* Main 2-Column Section with equal height stretching */}
                <div className="mt-12 grid grid-cols-1 gap-8 lg:grid-cols-12 lg:items-stretch">
                    {/* Left Column: Radio Cards List dividing full height evenly */}
                    <div className="flex flex-col justify-between gap-3.5 lg:col-span-6 h-full">
                        {PLANS.map((plan) => {
                            const isSelected = selectedKind === plan.kind
                            const priceInfo = getDisplayPrice(plan)

                            return (
                                <div
                                    key={plan.kind}
                                    onClick={() => setSelectedKind(plan.kind)}
                                    className={`group relative flex flex-1 cursor-pointer items-center justify-between rounded-2xl border p-4 sm:p-5 transition-all duration-200 ${isSelected
                                        ? "border-foreground bg-card shadow-sm ring-1 ring-foreground/10"
                                        : "border-border/70 bg-card/60 hover:border-border hover:bg-card"
                                        }`}
                                >
                                    <div className="flex items-center gap-4">
                                        {/* Custom Radio Circle Button */}
                                        <div
                                            className={`flex size-5 shrink-0 items-center justify-center rounded-full border transition-all ${isSelected
                                                ? "border-foreground bg-foreground text-background"
                                                : "border-muted-foreground/40 bg-background group-hover:border-muted-foreground"
                                                }`}
                                        >
                                            {isSelected && <div className="size-2 rounded-full bg-background" />}
                                        </div>

                                        <div>
                                            <div className="flex items-center gap-2">
                                                <span className="font-semibold text-base text-foreground">
                                                    {plan.name}
                                                </span>
                                            </div>
                                            {plan.badge && (
                                                <span className="mt-1 inline-block rounded-full bg-muted/80 px-2.5 py-0.5 text-[11px] font-medium text-muted-foreground">
                                                    {plan.kind === "flash" ? "Fixed Rate" : plan.badge}
                                                </span>
                                            )}
                                        </div>
                                    </div>

                                    {/* Price info on right of card */}
                                    <div className="text-right">
                                        <span className="font-bold text-xl sm:text-2xl text-foreground">
                                            {priceInfo.amount}
                                        </span>
                                        <span className="ml-1 text-xs text-muted-foreground">
                                            {priceInfo.period}
                                        </span>
                                    </div>
                                </div>
                            )
                        })}
                    </div>

                    {/* Right Column: Fixed Height Container to prevent layout shift */}
                    <div className="flex flex-col justify-between rounded-2xl border border-border/80 bg-card p-6 sm:p-8 shadow-xs lg:col-span-6 h-full min-h-127.5">
                        <div className="flex flex-col flex-1 justify-start">
                            <div className="flex items-center justify-between border-b pb-4 shrink-0">
                                <div>
                                    <h3 className="font-semibold text-lg text-foreground">
                                        Includes:
                                    </h3>
                                    <p className="text-xs text-muted-foreground">
                                        Features included in the {selectedPlan.name} plan
                                    </p>
                                </div>
                                <selectedPlan.Icon className="size-6 text-muted-foreground" />
                            </div>

                            {/* Counter Increaser/Decreaser for Custom Plan */}
                            {selectedKind === "custom" && (
                                <div className="mt-4 rounded-xl border border-border/80 bg-muted/30 p-3.5 shrink-0">
                                    <div className="flex items-center justify-between gap-3">
                                        <div>
                                            <label className="text-xs font-semibold text-foreground block">
                                                Set number of fixed events
                                            </label>
                                            <span className="text-[11px] text-muted-foreground">
                                                ${billingCycle === "yearly" ? selectedPlan.yearlyPrice : selectedPlan.monthlyPrice} per event / month
                                            </span>
                                        </div>

                                        {/* Counter Buttons */}
                                        <div className="flex items-center gap-1.5 rounded-lg border bg-background p-1 shadow-2xs">
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                className="size-7 rounded-md"
                                                onClick={() => setCustomEventCount((c) => Math.max(1, c - 1))}
                                                disabled={customEventCount <= 1}
                                            >
                                                <Minus className="size-3" />
                                            </Button>
                                            <span className="w-7 text-center text-xs font-bold text-foreground">
                                                {customEventCount}
                                            </span>
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                className="size-7 rounded-md"
                                                onClick={() => setCustomEventCount((c) => Math.min(100, c + 1))}
                                                disabled={customEventCount >= 100}
                                            >
                                                <Plus className="size-3" />
                                            </Button>
                                        </div>
                                    </div>

                                    <div className="mt-2.5 flex items-center justify-between border-t border-border/60 pt-2 text-[11px]">
                                        <span className="text-muted-foreground">Calculated total:</span>
                                        <span className="font-semibold text-foreground">
                                            {customEventCount} events × ${billingCycle === "yearly" ? selectedPlan.yearlyPrice : selectedPlan.monthlyPrice} = {selectedPriceInfo.amount}/month
                                        </span>
                                    </div>
                                </div>
                            )}

                            {/* Features List */}
                            <ul className="mt-5 space-y-3 text-sm text-foreground/90 flex-1">
                                {selectedPlan.features(customEventCount).map((feature) => (
                                    <li key={feature} className="flex items-start gap-3">
                                        <div className="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-600 dark:text-emerald-400">
                                            <Check className="size-3 stroke-3" />
                                        </div>
                                        <span>{feature}</span>
                                    </li>
                                ))}
                            </ul>
                        </div>

                        {/* CTA Button pinned cleanly to bottom */}
                        <div className="mt-6 pt-4 border-t border-border/40 shrink-0">
                            <Button
                                size="lg"
                                className="w-full rounded-full py-6 font-semibold bg-foreground text-background hover:bg-foreground/90 transition-colors shadow-xs"
                                render={<Link href="/login" />}
                            >
                                Choose Plan
                            </Button>
                        </div>
                    </div>
                </div>

                {/* Bottom Comparison Table */}
                <div className="mt-20 rounded-2xl border border-border/80 bg-card p-6 sm:p-8 shadow-2xs">
                    <div className="mb-6">
                        <h3 className="font-heading text-xl font-bold tracking-tight text-foreground sm:text-2xl">
                            Compare Plan Features
                        </h3>
                        <p className="mt-1 text-sm text-muted-foreground">
                            Detailed side-by-side comparison across all 4 plans.
                        </p>
                    </div>

                    <div className="overflow-x-auto">
                        <table className="w-full min-w-180 border-collapse text-sm">
                            <thead>
                                <tr className="border-b text-muted-foreground">
                                    <th className="py-3.5 pr-4 text-left font-medium">Feature</th>
                                    {PLANS.map((p) => (
                                        <th
                                            key={p.kind}
                                            className={`px-4 py-3.5 text-center font-semibold transition-colors ${selectedKind === p.kind
                                                ? "bg-muted/60 text-foreground rounded-t-lg"
                                                : "text-foreground"
                                                }`}
                                        >
                                            {p.name}
                                        </th>
                                    ))}
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-border/60">
                                {/* Row 1: Event Limit */}
                                <tr>
                                    <td className="py-4 pr-4 font-medium text-foreground">
                                        Event Creation Limit
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "flash" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        1 Event (Fixed)
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "base" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        1 Active Event
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "custom" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Custom ({customEventCount} Events)
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "unlimited" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Unlimited Events
                                    </td>
                                </tr>

                                {/* Row 2: Billing Type */}
                                <tr>
                                    <td className="py-4 pr-4 font-medium text-foreground">
                                        Billing Options
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "flash" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Fixed / One-Time ($29)
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "base" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Monthly or Yearly
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "custom" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Monthly or Yearly
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "unlimited" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Monthly or Yearly
                                    </td>
                                </tr>

                                {/* Row 3: Digital ID Cards */}
                                <tr>
                                    <td className="py-4 pr-4 font-medium text-foreground">
                                        Digital ID Cards (PDF + QR)
                                    </td>
                                    {PLANS.map((p) => (
                                        <td
                                            key={p.kind}
                                            className={`px-4 py-4 text-center text-emerald-600 dark:text-emerald-400 font-medium ${selectedKind === p.kind ? "bg-muted/40" : ""
                                                }`}
                                        >
                                            ✓ Included
                                        </td>
                                    ))}
                                </tr>

                                {/* Row 4: Scanner Devices */}
                                <tr>
                                    <td className="py-4 pr-4 font-medium text-foreground">
                                        Scanner & Door Check-in
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "flash" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        1 Scanner Pair
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "base" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Unlimited Scanners
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "custom" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Unlimited Scanners
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "unlimited" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Unlimited Scanners
                                    </td>
                                </tr>

                                {/* Row 5: Attendance Analytics */}
                                <tr>
                                    <td className="py-4 pr-4 font-medium text-foreground">
                                        Attendance Analytics & Exports
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "flash" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Basic Logs
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "base" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Standard Reports
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "custom" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Advanced Analytics + CSV
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "unlimited" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Full Analytics + Chart Downloads
                                    </td>
                                </tr>

                                {/* Row 6: Sub-events */}
                                <tr>
                                    <td className="py-4 pr-4 font-medium text-foreground">
                                        Sub-events & Multi-day
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "flash" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Single-day only
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "base" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        ✓ Multi-day
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "custom" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        ✓ Multi-day & Sub-events
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "unlimited" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        ✓ Multi-day & Sub-events
                                    </td>
                                </tr>

                                {/* Row 7: Support Level */}
                                <tr>
                                    <td className="py-4 pr-4 font-medium text-foreground">
                                        Support Level
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "flash" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Email Support
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "base" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Standard Support
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "custom" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        Priority Support
                                    </td>
                                    <td className={`px-4 py-4 text-center text-muted-foreground ${selectedKind === "unlimited" ? "bg-muted/40 font-medium text-foreground" : ""}`}>
                                        24/7 Dedicated Support
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    )
}
