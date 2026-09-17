"use client";

import * as React from "react";
import { CheckCircle2, Copy } from "lucide-react";
import { toast } from "sonner";

import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import type { SubmitApplicationResponse } from "../../schema/publicapply.types";

interface ApplySuccessCardProps {
    submission: SubmitApplicationResponse;
    eventName?: string;
    email: string;
}

export function ApplySuccessCard({
    submission,
    eventName,
    email,
}: ApplySuccessCardProps) {
    const [copiedUserId, setCopiedUserId] = React.useState(false);

    const handleCopyUserId = async () => {
        if (!submission.user_id) return;
        try {
            await navigator.clipboard.writeText(submission.user_id);
            setCopiedUserId(true);
            toast.success("User ID copied to clipboard");
            setTimeout(() => setCopiedUserId(false), 2000);
        } catch {
            toast.error("Failed to copy User ID");
        }
    };

    return (
        <div className="flex min-h-screen w-full items-center justify-center p-4 bg-muted/20">
            <Card className="max-w-lg w-full text-center p-6 sm:p-8 space-y-6">
                <div className="mx-auto flex size-14 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-500">
                    <CheckCircle2 className="size-8" />
                </div>

                <div className="space-y-1.5">
                    <Badge
                        variant="outline"
                        className="border-emerald-500/30 bg-emerald-500/10 text-emerald-600 font-medium text-xs"
                    >
                        Registration Confirmed
                    </Badge>
                    <h2 className="text-xl font-bold text-foreground">
                        You are successfully registered!
                    </h2>
                    <p className="text-xs text-muted-foreground">
                        Thank you for submitting your application for{" "}
                        <span className="font-semibold text-foreground">{eventName || "the event"}</span>.
                    </p>
                </div>

                {/* Assigned User ID display */}
                <div className="rounded-xl border bg-muted/30 p-4 space-y-2 text-left">
                    <span className="text-[11px] font-medium text-muted-foreground block">
                        Your Unique Applicant ID
                    </span>
                    <div className="flex items-center justify-between gap-2 bg-card border rounded-lg p-2.5">
                        <span className="font-mono text-sm font-bold text-primary tracking-wide">
                            {submission.user_id}
                        </span>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleCopyUserId}
                            className="gap-2"
                        >
                            {copiedUserId ? (
                                <CheckCircle2 className="size-3.5 text-emerald-500" />
                            ) : (
                                <Copy className="size-3.5" />
                            )}
                            <span>{copiedUserId ? "Copied" : "Copy ID"}</span>
                        </Button>
                    </div>
                    <p className="text-[10px] text-muted-foreground">
                        Please save this User ID. It will be printed on your official event badge / identity card.
                    </p>
                </div>

                <div className="text-xs text-muted-foreground border-t pt-4">
                    A confirmation record has been created with email{" "}
                    <span className="font-medium text-foreground">{email}</span>.
                </div>
            </Card>
        </div>
    );
}
