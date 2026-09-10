"use client";

import { useMemo, useState, useTransition } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { BILLING_CYCLE_LABEL, formatPlanPrice } from "@/components/billing/plan-card";
import { apiClient } from "@/react-query/client";
import { queryKeys } from "@/react-query/query-keys";
import type { BillingCycle, Plan, PlanKind } from "@/schema/plans.types";

const KIND_LABEL: Record<PlanKind, string> = {
    flash: "Flash",
    base: "Base",
    custom: "Custom",
    unlimited: "Unlimited",
};

const KIND_DESCRIPTION: Record<PlanKind, string> = {
    flash: "One single-day event, paid once.",
    base: "One event, for as long as you keep paying.",
    custom: "Pick how many events you need — buy more any time.",
    unlimited: "Unlimited events, no counting.",
};

export function SubscribeDialog({
    plans,
    trigger,
    title = "Choose a plan",
}: {
    plans: Plan[];
    trigger: React.ReactElement;
    title?: string;
}) {
    const queryClient = useQueryClient();
    const [open, setOpen] = useState(false);
    const [kind, setKind] = useState<PlanKind | null>(null);
    const [cycle, setCycle] = useState<BillingCycle | null>(null);
    const [quantity, setQuantity] = useState(1);
    const [error, setError] = useState<string | null>(null);
    const [isPending, startTransition] = useTransition();

    const kinds = useMemo(
        () => Array.from(new Set(plans.map((p) => p.kind))),
        [plans]
    );
    const cyclesForKind = useMemo(
        () => plans.filter((p) => p.kind === kind).map((p) => p.billing_cycle),
        [plans, kind]
    );
    const selectedPlan = plans.find(
        (p) => p.kind === kind && (kind === "flash" ? true : p.billing_cycle === cycle)
    );
    const isCustom = kind === "custom";
    const total =
        selectedPlan &&
        (isCustom
            ? selectedPlan.per_event_amount != null
                ? selectedPlan.per_event_amount * Math.max(1, quantity)
                : null
            : (selectedPlan.amount ?? null));

    function reset() {
        setKind(null);
        setCycle(null);
        setQuantity(1);
        setError(null);
    }

    function selectKind(next: PlanKind) {
        setKind(next);
        setCycle(next === "flash" ? "one_time" : null);
        setQuantity(1);
        setError(null);
    }

    function confirm() {
        if (!selectedPlan) return;
        setError(null);
        startTransition(async () => {
            try {
                await apiClient.post("/plans/purchase", {
                    plan_code: selectedPlan.code,
                    event_quantity: isCustom ? quantity : undefined,
                });
            } catch (err) {
                setError(err instanceof Error ? err.message : "Something went wrong.");
                return;
            }
            queryClient.invalidateQueries({ queryKey: queryKeys.orgSubscriptions() });
            queryClient.invalidateQueries({ queryKey: queryKeys.plans() });
            setOpen(false);
            reset();
        });
    }

    return (
        <Dialog
            open={open}
            onOpenChange={(next) => {
                setOpen(next);
                if (!next) reset();
            }}
        >
            <DialogTrigger render={trigger} />
            <DialogContent className="sm:max-w-lg">
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                    <DialogDescription>
                        Buying is additive — if you already have a plan, this adds to it.
                    </DialogDescription>
                </DialogHeader>

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
                                    (kind === k
                                        ? "border-primary ring-1 ring-primary"
                                        : "hover:bg-muted")
                                }
                            >
                                <span className="font-medium">{KIND_LABEL[k]}</span>
                                <span className="text-xs text-muted-foreground">
                                    {KIND_DESCRIPTION[k]}
                                </span>
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
                                        (cycle === c
                                            ? "border-primary ring-1 ring-primary"
                                            : "hover:bg-muted")
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

                {selectedPlan && total != null && (
                    <div className="flex items-center justify-between rounded-lg bg-muted/50 p-3 text-sm">
                        <span>Total due today</span>
                        <span className="font-medium">
                            {new Intl.NumberFormat("en-IN", {
                                style: "currency",
                                currency: selectedPlan.currency,
                                maximumFractionDigits: 0,
                            }).format(total / 100)}
                        </span>
                    </div>
                )}

                <DialogFooter>
                    <Button
                        disabled={!selectedPlan || (kind !== "flash" && !cycle) || isPending}
                        onClick={confirm}
                        className="w-full"
                    >
                        {isPending ? "Confirming…" : "Pay & confirm"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}