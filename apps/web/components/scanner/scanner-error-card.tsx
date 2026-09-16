"use client";

import * as React from "react";
import { RefreshCw, ShieldAlert } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

interface ScannerErrorCardProps {
    errorMessage: string;
    onScanAgain: () => void;
}

export function ScannerErrorCard({
    errorMessage,
    onScanAgain,
}: ScannerErrorCardProps) {
    return (
        <Card className="w-full max-w-md border-destructive/40 bg-destructive/5 shadow-sm">
            <CardHeader className="text-center pb-3">
                <div className="mx-auto flex size-12 items-center justify-center rounded-2xl bg-destructive/15 text-destructive mb-2">
                    <ShieldAlert className="size-6" />
                </div>
                <CardTitle className="text-base font-bold text-destructive">
                    Admission Error
                </CardTitle>
                <CardDescription className="text-xs text-muted-foreground">
                    The scanned badge could not be admitted for this event.
                </CardDescription>
            </CardHeader>

            <CardContent className="space-y-4">
                <div className="rounded-xl border border-destructive/20 bg-background/80 p-4 text-center">
                    <p className="text-xs font-semibold text-destructive leading-relaxed">
                        {errorMessage}
                    </p>
                    <p className="mt-1 text-[11px] text-muted-foreground">
                        Ensure the attendee is registered for this event and your scanner terminal is authorized.
                    </p>
                </div>

                <div className="pt-2 flex flex-col gap-2">
                    <Button
                        type="button"
                        variant="destructive"
                        onClick={onScanAgain}
                        className="w-full gap-2 text-xs font-semibold h-10"
                    >
                        <RefreshCw className="size-4" />
                        <span>Scan Another Badge</span>
                    </Button>
                </div>
            </CardContent>
        </Card>
    );
}
