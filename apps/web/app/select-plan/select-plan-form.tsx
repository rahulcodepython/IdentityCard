"use client"

import { useMemo, useState, useTransition } from "react"
import { useRouter } from "next/navigation"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { BILLING_CYCLE_LABEL, formatPlanPrice } from "@/components/plan-card"
import { purchasePlan } from "@/lib/actions/checkout"
import type { BillingCycle, Plan, PlanKind } from "@/schema/plans.types"
import { useSessionStore } from "@/store/session.store"

const KIND_LABEL: Record<PlanKind, string> = {
    flash: "Flash",
    base: "Base",
    custom: "Custom",
    unlimited: "Unlimited",
}

const KIND_DESCRIPTION: Record<PlanKind, string> = {
    flash: "One single-day event, paid once.",
    base: "One event, for as long as you keep paying.",
    custom: "Pick how many events you need — buy more any time.",
    unlimited: "Unlimited events, no counting.",
}

export function SelectPlanForm({ plans, defaultName }: { plans: Plan[]; defaultName: string }) {
    const router = useRouter()
    const setToken = useSessionStore((s) => s.setToken)
    const [organizationName, setOrganizationName] = useState(`${defaultName}'s Organization`)
    const [kind, setKind] = useState<PlanKind | null>(null)
    const [cycle, setCycle] = useState<BillingCycle | null>(null)
    const [quantity, setQuantity] = useState(1)
    const [error, setError] = useState<string | null>(null)
    const [isPending, startTransition] = useTransition()

    const kinds = useMemo(() => Array.from(new Set(plans.map((p) => p.kind))), [plans])
    const cyclesForKind = useMemo(
        () => plans.filter((p) => p.kind === kind).map((p) => p.billing_cycle),
        [plans, kind]
    )
    const selectedPlan = plans.find(
        (p) => p.kind === kind && (kind === "flash" ? true : p.billing_cycle === cycle)
    )
    const isCustom = kind === "custom"

    function selectKind(next: PlanKind) {
        setKind(next)
        setCycle(next === "flash" ? "one_time" : null)
        setQuantity(1)
        setError(null)
    }

    function confirm() {
        if (!selectedPlan || !organizationName.trim()) return
        setError(null)
        startTransition(async () => {
            try {
                await purchasePlan({
                    planCode: selectedPlan.code,
                    eventQuantity: isCustom ? quantity : undefined,
                    organizationName,
                })
            } catch (err) {
                setError(err instanceof Error ? err.message : "Something went wrong.")
                return
            }
            // Org id/role changed — the store's token needs to reflect the
            // org this purchase just created (see components/session-provider.tsx's
            // doc comment: no second getSession() call, just a fresh token).
            const { authClient } = await import("@/lib/auth-client")
            const { decodeJwtPayload } = await import("@/lib/jwt")
            const { data } = await authClient.token()
            if (data?.token) {
                const decoded = decodeJwtPayload(data.token)
                setToken(data.token, decoded.organizationId, decoded.role)
            }
            router.push("/dashboard")
        })
    }

    return (
        <Card className="w-full max-w-lg">
            <CardContent className="flex flex-col gap-6 p-6">
                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="organization-name">Organization name</FieldLabel>
                        <Input
                            id="organization-name"
                            value={organizationName}
                            onChange={(e) => setOrganizationName(e.target.value)}
                            placeholder="e.g. Acme Events"
                        />
                        <FieldDescription>You can change this later from Settings.</FieldDescription>
                    </Field>
                </FieldGroup>

                <div className="flex flex-col gap-2">
                    <Label>Plan</Label>
                    <div className="grid grid-cols-2 gap-2">
                        {kinds.map((k) => (
                            <button
                                key={k}
                                type="button"
                                onClick={() => selectKind(k)}
                                className={
                                    "flex flex-col items-start gap-0.5 rounded-lg border p-3 text-left text-sm transition-colors " +
                                    (kind === k ? "border-primary ring-1 ring-primary" : "hover:bg-muted")
                                }
                            >
                                <span className="font-medium">{KIND_LABEL[k]}</span>
                                <span className="text-xs text-muted-foreground">{KIND_DESCRIPTION[k]}</span>
                            </button>
                        ))}
                    </div>
                </div>

                {kind && kind !== "flash" && (
                    <div className="flex flex-col gap-2">
                        <Label>Billing cycle</Label>
                        <div className="grid grid-cols-2 gap-2">
                            {cyclesForKind.map((c) => (
                                <button
                                    key={c}
                                    type="button"
                                    onClick={() => setCycle(c)}
                                    className={
                                        "rounded-lg border p-2.5 text-center text-sm transition-colors " +
                                        (cycle === c ? "border-primary ring-1 ring-primary" : "hover:bg-muted")
                                    }
                                >
                                    {BILLING_CYCLE_LABEL[c]}
                                </button>
                            ))}
                        </div>
                    </div>
                )}

                {isCustom && selectedPlan && (
                    <div className="flex flex-col gap-1.5">
                        <Label htmlFor="event-quantity">Number of events</Label>
                        <Input
                            id="event-quantity"
                            type="number"
                            min={1}
                            value={quantity}
                            onChange={(e) => setQuantity(Math.max(1, Number(e.target.value) || 1))}
                        />
                        <p className="text-xs text-muted-foreground">
                            {formatPlanPrice(selectedPlan)} × {quantity}
                        </p>
                    </div>
                )}

                {error && <p className="text-sm text-destructive">{error}</p>}

                <Button
                    disabled={!selectedPlan || (kind !== "flash" && !cycle) || !organizationName.trim() || isPending}
                    onClick={confirm}
                    className="w-full"
                >
                    {isPending ? "Setting up…" : "Subscribe & create organization"}
                </Button>
            </CardContent>
        </Card>
    )
}
