"use client";

import * as React from "react";
import {
    Calendar,
    Check,
    Copy,
    ExternalLink,
    KeyRound,
    Loader2,
    Plus,
    Smartphone,
} from "lucide-react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
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
import { useCreateDeviceMutation } from "@/query-hooks/devices.api";
import type { Device } from "@/schema/devices.types";

interface CreateDeviceDialogProps {
    open?: boolean;
    onOpenChange?: (open: boolean) => void;
    trigger?: React.ReactNode;
}

export function CreateDeviceDialog({
    open: controlledOpen,
    onOpenChange: setControlledOpen,
    trigger,
}: CreateDeviceDialogProps) {
    const [internalOpen, setInternalOpen] = React.useState(false);
    const isOpen = controlledOpen !== undefined ? controlledOpen : internalOpen;
    const setIsOpen = setControlledOpen !== undefined ? setControlledOpen : setInternalOpen;

    const [name, setName] = React.useState("");
    const [isUnlimitedTTL, setIsUnlimitedTTL] = React.useState(true);
    const [expiresAt, setExpiresAt] = React.useState("");
    const [createdDevice, setCreatedDevice] = React.useState<Device | null>(null);

    const createGlobalMutation = useCreateDeviceMutation();
    const isSubmitting = createGlobalMutation.isPending;

    const handleReset = () => {
        setName("");
        setIsUnlimitedTTL(true);
        setExpiresAt("");
        setCreatedDevice(null);
    };

    const handleDialogChange = (open: boolean) => {
        setIsOpen(open);
        if (!open) {
            handleReset();
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = name.trim();
        if (!trimmed) return;

        const targetExpires = !isUnlimitedTTL && expiresAt ? new Date(expiresAt).toISOString() : null;

        try {
            const dev = await createGlobalMutation.mutateAsync({
                name: trimmed,
                expires_at: targetExpires,
            });
            setCreatedDevice(dev);
        } catch {
            // Handled by mutation toast
        }
    };

    const handleCopy = (text: string) => {
        navigator.clipboard.writeText(text);
        toast.success("Copied to clipboard");
    };

    return (
        <Dialog open={isOpen} onOpenChange={handleDialogChange}>
            {
                trigger ? <DialogTrigger render={trigger as React.ReactElement} /> : null
            }

            <DialogContent className="sm:max-w-md p-0 overflow-hidden flex flex-col">
                <DialogHeader className="p-5 pb-4 border-b bg-muted/20">
                    <div className="flex items-center gap-2.5">
                        <div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                            <Smartphone className="size-4" />
                        </div>
                        <div>
                            <DialogTitle className="text-sm font-semibold">
                                {
                                    createdDevice ? "Device Added & Ready to Pair" : "Add Scanner Device"
                                }
                            </DialogTitle>
                            <DialogDescription className="text-xs text-muted-foreground mt-0.5">
                                {
                                    createdDevice
                                        ? "Enter this 6-digit PIN on the scanner device to pair it."
                                        : "Create a new terminal record and generate a 5-minute verification PIN."
                                }
                            </DialogDescription>
                        </div>
                    </div>
                </DialogHeader>

                {/* State A: Created successfully, display PIN & Instructions */}
                {
                    createdDevice ? <div className="p-5 space-y-4">
                        <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-4 text-center space-y-3">
                            <span className="text-xs text-muted-foreground font-medium">
                                5-Minute Verification PIN
                            </span>

                            <div className="flex items-center justify-center gap-3">
                                <span className="font-mono text-3xl font-bold tracking-widest text-foreground">
                                    {createdDevice.pin}
                                </span>
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => handleCopy(createdDevice.pin || "")}
                                    className="gap-1 text-xs"
                                >
                                    <Copy className="size-3.5" />
                                    <span>Copy</span>
                                </Button>
                            </div>

                            <p className="text-[11px] text-muted-foreground">
                                PIN expires in 5 minutes. If it expires, you can regenerate it from the dashboard.
                            </p>
                        </div>

                        <div className="rounded-lg border bg-muted/30 p-3 space-y-2 text-xs">
                            <div className="flex items-center justify-between">
                                <span className="text-muted-foreground">Device Name:</span>
                                <span className="font-semibold text-foreground">{createdDevice.name}</span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-muted-foreground">Public Pairing URL:</span>
                                <a
                                    href="/devices/pair"
                                    target="_blank"
                                    rel="noreferrer"
                                    className="text-primary hover:underline flex items-center gap-1 font-mono text-[11px]"
                                >
                                    <span>/devices/pair</span>
                                    <ExternalLink className="size-3" />
                                </a>
                            </div>
                        </div>

                        <DialogFooter className="pt-2">
                            <Button
                                type="button"
                                variant="default"
                                onClick={() => handleDialogChange(false)}
                                className="w-full text-xs font-semibold"
                            >
                                Done
                            </Button>
                        </DialogFooter>
                    </div> : null
                }

                {/* State B: Device creation form */}
                {
                    !createdDevice ? <form onSubmit={handleSubmit} className="p-5 space-y-4">
                        <div className="space-y-2">
                            <Label className="text-xs font-medium">Device Name</Label>
                            <Input
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                placeholder="e.g. Main Gate Entrance Scanner A"
                                className="text-xs h-9"
                                required
                                autoFocus
                            />
                            <p className="text-[11px] text-muted-foreground">
                                A descriptive name to identify this scanner in your dashboard.
                            </p>
                        </div>

                        {/* TTL / Expiration Section */}
                        <div className="space-y-2 pt-1 border-t">
                            <Label className="text-xs font-medium flex items-center gap-1.5">
                                <Calendar className="size-3.5 text-muted-foreground" />
                                <span>Device Session Lifetime (TTL)</span>
                            </Label>

                            <div className="space-y-2">
                                <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer select-none">
                                    <Checkbox
                                        checked={isUnlimitedTTL}
                                        onCheckedChange={(checked) => setIsUnlimitedTTL(!!checked)}
                                    />
                                    <span>Unlimited TTL (stays verified until manually unpaired)</span>
                                </label>

                                {
                                    !isUnlimitedTTL ? <div className="space-y-1 pt-1">
                                        <Input
                                            type="datetime-local"
                                            value={expiresAt}
                                            onChange={(e) => setExpiresAt(e.target.value)}
                                            className="text-xs h-9 font-mono"
                                            required={!isUnlimitedTTL}
                                        />
                                        <p className="text-[11px] text-muted-foreground">
                                            After this time, the device will expire and require re-verification.
                                        </p>
                                    </div> : null
                                }
                            </div>
                        </div>

                        <DialogFooter className="pt-3 gap-2 sm:gap-0">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => handleDialogChange(false)}
                                className="text-xs"
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                disabled={!name.trim() || isSubmitting || (!isUnlimitedTTL && !expiresAt)}
                                className="text-xs font-semibold gap-1.5"
                            >
                                {
                                    isSubmitting ? <>
                                        <Loader2 className="size-3.5 animate-spin" />
                                        <span>Creating Device...</span>
                                    </> : <>
                                        <Plus className="size-3.5" />
                                        <span>Create & Generate PIN</span>
                                    </>
                                }
                            </Button>
                        </DialogFooter>
                    </form> : null
                }
            </DialogContent>
        </Dialog>
    );
}
