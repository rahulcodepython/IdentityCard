import * as React from "react";
import { CheckCircle2, Fingerprint, RefreshCw, ScanLine, ShieldCheck, Smartphone } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { Device } from "@/schema/devices.types";

interface PairSessionCardProps {
    device: Device;
    onDisconnect: () => void;
    onEnrollBiometrics?: () => void;
    isBiometricSupported: boolean;
    isEnrollingBiometrics?: boolean;
}

export function PairSessionCard({
    device,
    onDisconnect,
    onEnrollBiometrics,
    isBiometricSupported,
    isEnrollingBiometrics,
}: PairSessionCardProps) {
    return (
        <Card className="w-full max-w-md shadow-sm border-border/80">
            <CardHeader className="text-center pb-4">
                <div className="mx-auto flex size-12 items-center justify-center rounded-2xl bg-emerald-500/10 text-emerald-600 mb-2">
                    <ShieldCheck className="size-6" />
                </div>
                <CardTitle className="text-base font-bold text-foreground">
                    Scanner Terminal Paired
                </CardTitle>
                <CardDescription className="text-xs">
                    This browser is authorized to scan tickets and check in attendees.
                </CardDescription>
            </CardHeader>

            <CardContent className="space-y-4 text-xs">
                <div className="rounded-xl border bg-muted/30 p-3.5 space-y-2.5">
                    <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Terminal Name</span>
                        <span className="font-semibold text-foreground">{device.name}</span>
                    </div>
                    <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Hardware Device</span>
                        <span className="font-medium text-foreground">{device.actual_name || "Web Browser"}</span>
                    </div>
                    <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Pairing Status</span>
                        <Badge variant="default" className="bg-emerald-600 hover:bg-emerald-600 text-[10px] gap-1">
                            <CheckCircle2 className="size-3" />
                            <span>Active Session</span>
                        </Badge>
                    </div>
                    <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Biometrics (Passkey)</span>
                        {device.is_biometric_enrolled ? (
                            <Badge variant="default" className="bg-emerald-600 hover:bg-emerald-600 text-[10px] gap-1">
                                <Fingerprint className="size-3" />
                                <span>Enrolled</span>
                            </Badge>
                        ) : (
                            <span className="text-[11px] text-muted-foreground italic">Not Enrolled</span>
                        )}
                    </div>
                </div>

                {!device.is_biometric_enrolled && isBiometricSupported && onEnrollBiometrics && (
                    <Button
                        type="button"
                        variant="outline"
                        onClick={onEnrollBiometrics}
                        disabled={isEnrollingBiometrics}
                        className="w-full gap-2 text-xs font-medium"
                    >
                        <Fingerprint className="size-4 text-primary" />
                        <span>Enable Face ID / Touch ID Re-auth</span>
                    </Button>
                )}

                <div className="pt-2 flex flex-col gap-2">
                    <Button
                        type="button"
                        variant="default"
                        onClick={() => window.location.href = "/devices/scan"}
                        className="w-full gap-2 text-xs font-semibold h-10"
                    >
                        <ScanLine className="size-4" />
                        <span>Open Scanner Terminal</span>
                    </Button>

                    <Button
                        type="button"
                        variant="ghost"
                        onClick={onDisconnect}
                        className="w-full text-xs text-muted-foreground hover:text-destructive"
                    >
                        Disconnect / Unpair Terminal
                    </Button>
                </div>
            </CardContent>
        </Card>
    );
}
