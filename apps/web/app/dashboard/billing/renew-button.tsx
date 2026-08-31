"use client";

import { Button } from "@/components/ui/button";
import { useRenewPlanMutation } from "@/query-hooks/plans.api";

export function RenewButton({ lineageRootId }: { lineageRootId: string }) {
    const renewMutation = useRenewPlanMutation();

    return (
        <div className="flex flex-col items-end gap-1">
            <Button
                size="sm"
                disabled={renewMutation.isPending}
                onClick={async () => {
                    await renewMutation.execute({ lineageRootId });
                }}
            >
                {renewMutation.isPending ? "Renewing…" : "Renew now"}
            </Button>
            {renewMutation.error && (
                <p className="text-xs text-destructive">{renewMutation.error.message}</p>
            )}
        </div>
    );
}