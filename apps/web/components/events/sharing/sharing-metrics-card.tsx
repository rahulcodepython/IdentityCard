"use client";

import * as React from "react";
import { AlertCircle, Clock, Globe, Lock, Users } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import type { EventSharingResponse } from "@/schema/eventsharing.types";

interface SharingMetricsCardProps {
    sharing?: EventSharingResponse;
}

export function SharingMetricsCard({ sharing }: SharingMetricsCardProps) {
    const isLive = sharing?.event_form?.status === "live";
    const totalApplicants = sharing?.total_applicants ?? 0;
    const maxApplicants = sharing?.event_form?.max_applicants ?? -1;
    const isFull = maxApplicants !== -1 && totalApplicants >= maxApplicants;
    const isExpired = sharing?.event_form?.expires_at
        ? new Date(sharing.event_form.expires_at).getTime() < Date.now()
        : false;

    return (
        <Card className="shadow-xs">
            <CardHeader className="pb-3">
                <CardTitle className="text-sm font-semibold flex items-center gap-2">
                    <Users className="size-4 text-primary" />
                    <span>Registration Status</span>
                </CardTitle>
                <CardDescription className="text-xs">
                    Current intake and status overview.
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
                <div className="flex items-center justify-between p-2.5 rounded-lg border bg-muted/20 text-xs">
                    <span className="text-muted-foreground">Status:</span>
                    {isLive ? (
                        <Badge
                            variant="outline"
                            className="gap-1 border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-[10px]"
                        >
                            <Globe className="size-2.5" />
                            <span>Accepting Responses</span>
                        </Badge>
                    ) : (
                        <Badge
                            variant="outline"
                            className="gap-1 border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400 text-[10px]"
                        >
                            <Lock className="size-2.5" />
                            <span>Form Paused (Waiting)</span>
                        </Badge>
                    )}
                </div>

                <div className="flex items-center justify-between p-2.5 rounded-lg border bg-muted/20 text-xs">
                    <span className="text-muted-foreground">Intake:</span>
                    <span className="font-semibold text-foreground">
                        {totalApplicants}{" "}
                        <span className="text-muted-foreground font-normal">
                            / {maxApplicants === -1 ? "Unlimited" : `${maxApplicants} max`}
                        </span>
                    </span>
                </div>

                {sharing?.event_form?.expires_at && (
                    <div className="flex items-center justify-between p-2.5 rounded-lg border bg-muted/20 text-xs">
                        <span className="text-muted-foreground">Closes At:</span>
                        <span className="font-mono text-[11px] text-foreground flex items-center gap-1">
                            <Clock className="size-3 text-muted-foreground" />
                            {new Date(sharing.event_form.expires_at).toLocaleString()}
                        </span>
                    </div>
                )}

                {isFull && (
                    <div className="flex items-start gap-2 p-2.5 rounded-lg border border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-400 text-xs">
                        <AlertCircle className="size-4 shrink-0 mt-0.5" />
                        <span>Registration capacity has been reached. New applicants will be blocked.</span>
                    </div>
                )}

                {isExpired && (
                    <div className="flex items-start gap-2 p-2.5 rounded-lg border border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-400 text-xs">
                        <AlertCircle className="size-4 shrink-0 mt-0.5" />
                        <span>The registration deadline has passed.</span>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
