"use client";

import * as React from "react";
import {
    AlertCircle,
    Calendar,
    Check,
    Clock,
    Lock,
    SlidersHorizontal,
    Users,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { EventFormDetails } from "@/schema/eventform.types";

interface EventFormSettingsDialogProps {
    eventForm: EventFormDetails;
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSaveSettings: (settings: {
        max_applicants: number;
        expires_at: string;
    }) => Promise<void>;
    isSaving?: boolean;
    isLockPrompt?: boolean;
    onProceedToLock?: () => void;
}

export function EventFormSettingsDialog({
    eventForm,
    open,
    onOpenChange,
    onSaveSettings,
    isSaving = false,
    isLockPrompt = false,
    onProceedToLock,
}: EventFormSettingsDialogProps) {
    const isLocked = eventForm.is_locked;

    // Convert ISO string to YYYY-MM-DDTHH:MM for datetime-local
    const formatForDatetimeLocal = (isoString?: string) => {
        if (!isoString) return "";
        try {
            const d = new Date(isoString);
            if (isNaN(d.getTime())) return "";
            const year = d.getFullYear();
            const month = String(d.getMonth() + 1).padStart(2, "0");
            const day = String(d.getDate()).padStart(2, "0");
            const hours = String(d.getHours()).padStart(2, "0");
            const minutes = String(d.getMinutes()).padStart(2, "0");
            return `${year}-${month}-${day}T${hours}:${minutes}`;
        } catch {
            return "";
        }
    };

    const [isUnlimited, setIsUnlimited] = React.useState<boolean>(
        eventForm.max_applicants === -1
    );
    const [maxApplicants, setMaxApplicants] = React.useState<number>(
        eventForm.max_applicants === -1 ? 100 : eventForm.max_applicants
    );
    const [expiresAt, setExpiresAt] = React.useState<string>(() =>
        formatForDatetimeLocal(eventForm.expires_at)
    );

    // Sync when dialog opens or eventForm changes
    React.useEffect(() => {
        if (open) {
            setIsUnlimited(eventForm.max_applicants === -1);
            setMaxApplicants(
                eventForm.max_applicants === -1 ? 100 : eventForm.max_applicants
            );
            setExpiresAt(formatForDatetimeLocal(eventForm.expires_at));
        }
    }, [open, eventForm]);

    const hasChanges = React.useMemo(() => {
        const currentMax = isUnlimited ? -1 : maxApplicants;
        if (currentMax !== eventForm.max_applicants) return true;
        const currentFormatted = formatForDatetimeLocal(eventForm.expires_at);
        if (expiresAt !== currentFormatted) return true;
        return false;
    }, [isUnlimited, maxApplicants, expiresAt, eventForm]);

    const isFutureDate = React.useMemo(() => {
        if (!expiresAt) return false;
        const time = new Date(expiresAt).getTime();
        return !isNaN(time) && time > Date.now();
    }, [expiresAt]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (isLocked) return;

        const targetMax = isUnlimited ? -1 : Number(maxApplicants) || 1;
        const targetExpires = expiresAt
            ? new Date(expiresAt).toISOString()
            : eventForm.expires_at;

        await onSaveSettings({
            max_applicants: targetMax,
            expires_at: targetExpires,
        });

        if (isLockPrompt && onProceedToLock) {
            onOpenChange(false);
            onProceedToLock();
        } else {
            onOpenChange(false);
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md p-0 overflow-hidden flex flex-col">
                <DialogHeader className="p-5 pb-4 border-b bg-muted/20">
                    <div className="flex items-center gap-2.5">
                        <div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                            <SlidersHorizontal className="size-4" />
                        </div>
                        <div>
                            <DialogTitle className="text-sm font-semibold">
                                Registration Limits & Deadline
                            </DialogTitle>
                            <DialogDescription className="text-xs text-muted-foreground mt-0.5">
                                Set attendee capacity and registration expiration deadline.
                            </DialogDescription>
                        </div>
                    </div>
                </DialogHeader>

                <form onSubmit={handleSubmit} className="p-5 space-y-4">
                    {isLockPrompt && (
                        <div className="rounded-lg border border-primary/30 bg-primary/5 p-3 text-xs text-foreground flex items-start gap-2.5">
                            <AlertCircle className="size-4 text-primary shrink-0 mt-0.5" />
                            <div className="space-y-1">
                                <p className="font-semibold text-primary">
                                    Set Limits Before Locking
                                </p>
                                <p className="text-[11px] text-muted-foreground leading-relaxed">
                                    Please verify the maximum capacity and registration deadline. Once locked, the form becomes immutable and live for public registration.
                                </p>
                            </div>
                        </div>
                    )}

                    {/* Capacity Section */}
                    <div className="space-y-2">
                        <div className="flex items-center justify-between">
                            <Label className="text-xs font-medium flex items-center gap-1.5">
                                <Users className="size-3.5 text-muted-foreground" />
                                <span>Max Applicants (Capacity)</span>
                            </Label>
                            <span className="text-[11px] text-muted-foreground">
                                Currently registered:{" "}
                                <strong className="text-foreground">
                                    {eventForm.total_applicants}
                                </strong>
                            </span>
                        </div>

                        {isLocked ? (
                            <div className="text-xs font-semibold text-foreground p-2.5 rounded-md bg-muted/40 border flex items-center gap-2">
                                <Lock className="size-3.5 text-muted-foreground" />
                                <span>
                                    {eventForm.max_applicants === -1
                                        ? "Unlimited Attendees"
                                        : `${eventForm.max_applicants} Attendees Max`}
                                </span>
                            </div>
                        ) : (
                            <div className="space-y-2">
                                <Input
                                    type="number"
                                    min={1}
                                    disabled={isUnlimited || isLocked}
                                    value={isUnlimited ? "" : maxApplicants}
                                    placeholder={isUnlimited ? "Unlimited" : "e.g. 150"}
                                    onChange={(e) =>
                                        setMaxApplicants(
                                            Math.max(1, Number(e.target.value))
                                        )
                                    }
                                    className="text-xs h-9"
                                />
                                <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer select-none">
                                    <Checkbox
                                        checked={isUnlimited}
                                        disabled={isLocked}
                                        onCheckedChange={(checked) =>
                                            setIsUnlimited(!!checked)
                                        }
                                    />
                                    <span>Unlimited capacity (no applicant cap)</span>
                                </label>
                            </div>
                        )}
                    </div>

                    {/* Deadline Section */}
                    <div className="space-y-2">
                        <Label className="text-xs font-medium flex items-center gap-1.5">
                            <Calendar className="size-3.5 text-muted-foreground" />
                            <span>Registration Deadline</span>
                        </Label>

                        {isLocked ? (
                            <div className="text-xs font-semibold text-foreground p-2.5 rounded-md bg-muted/40 border flex items-center gap-2">
                                <Clock className="size-3.5 text-muted-foreground" />
                                <span>
                                    {new Date(eventForm.expires_at).toLocaleString()}
                                </span>
                            </div>
                        ) : (
                            <div className="space-y-1">
                                <Input
                                    type="datetime-local"
                                    disabled={isLocked}
                                    value={expiresAt}
                                    onChange={(e) => setExpiresAt(e.target.value)}
                                    className="text-xs h-9 font-mono"
                                    required
                                />
                                {!isFutureDate && expiresAt && (
                                    <p className="text-[11px] text-destructive font-medium">
                                        Deadline must be in the future.
                                    </p>
                                )}
                            </div>
                        )}
                    </div>

                    <DialogFooter className="pt-2 gap-2 sm:gap-0">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => onOpenChange(false)}
                            className="text-xs"
                        >
                            Cancel
                        </Button>

                        {!isLocked && (
                            <Button
                                type="submit"
                                disabled={
                                    isSaving ||
                                    !isFutureDate ||
                                    (!isUnlimited && maxApplicants < 1) ||
                                    (!isLockPrompt && !hasChanges)
                                }
                                className="text-xs font-semibold gap-1.5"
                            >
                                {isLockPrompt ? (
                                    <>
                                        <Lock className="size-3.5" />
                                        <span>
                                            {isSaving
                                                ? "Saving..."
                                                : "Save & Proceed to Lock"}
                                        </span>
                                    </>
                                ) : (
                                    <>
                                        <Check className="size-3.5" />
                                        <span>
                                            {isSaving ? "Saving..." : "Save Limits"}
                                        </span>
                                    </>
                                )}
                            </Button>
                        )}
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
