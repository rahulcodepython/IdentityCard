"use client";

import * as React from "react";
import { Globe, Lock } from "lucide-react";

import { Badge } from "@/components/ui/badge";

interface SharingHeaderProps {
    eventName?: string;
    isLive: boolean;
}

export function SharingHeader({ eventName, isLive }: SharingHeaderProps) {
    return (
        <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between px-1">
            <div>
                <div className="flex items-center gap-2">
                    <h1 className="text-lg font-bold text-foreground">
                        Event Sharing & Registration
                    </h1>
                    {isLive ? (
                        <Badge
                            variant="outline"
                            className="gap-1 border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-[10px]"
                        >
                            <Globe className="size-2.5" />
                            <span>Live</span>
                        </Badge>
                    ) : (
                        <Badge
                            variant="outline"
                            className="gap-1 border-muted bg-muted/40 text-muted-foreground text-[10px]"
                        >
                            <Lock className="size-2.5" />
                            <span>Waiting</span>
                        </Badge>
                    )}
                </div>
                <p className="text-xs text-muted-foreground">
                    Assign a registration form, set applicant quotas, and share the public link for{" "}
                    {eventName || "this event"}.
                </p>
            </div>
        </div>
    );
}
