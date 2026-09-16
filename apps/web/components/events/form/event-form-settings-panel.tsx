"use client";

import * as React from "react";
import { Calendar, Check, ChevronDown, ChevronUp, Clock, Lock, Users } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import type { EventFormDetails } from "@/schema/eventform.types";

interface EventFormSettingsPanelProps {
    eventForm: EventFormDetails;
    onSaveSettings: (settings: {
        max_applicants: number;
        expires_at: string;
    }) => Promise<void>;
    isSaving?: boolean;
    defaultOpen?: boolean;
}

export function EventFormSettingsPanel({
    eventForm,
    onSaveSettings,
    isSaving = false,
    defaultOpen = false,
}: EventFormSettingsPanelProps) {
    const isLocked = eventForm.is_locked;
    const [isOpen, setIsOpen] = React.useState<boolean>(defaultOpen);

    // Local form state
    const [isUnlimited, setIsUnlimited] = React.useState<boolean>(
        eventForm.max_applicants === -1
    );
    const [maxApplicants, setMaxApplicants] = React.useState<number>(
        eventForm.max_applicants === -1 ? 100 : eventForm.max_applicants
    );

    // Convert ISO string to format YYYY-MM-DDTHH:MM for datetime-local input
    const formatForDatetimeLocal = (isoString: string) => {
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

    const [expiresAt, setExpiresAt] = React.useState<string>(() =>
        formatForDatetimeLocal(eventForm.expires_at)
    );

    // Track if settings changed
    const hasChanges = React.useMemo(() => {
        const currentMax = isUnlimited ? -1 : maxApplicants;
        if (currentMax !== eventForm.max_applicants) return true;
        const currentFormatted = formatForDatetimeLocal(eventForm.expires_at);
        if (expiresAt !== currentFormatted) return true;
        return false;
    }, [isUnlimited, maxApplicants, expiresAt, eventForm]);

    const handleSave = async (e: React.FormEvent) => {
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
    };

    const formattedDeadline = React.useMemo(() => {
        try {
            const d = new Date(eventForm.expires_at);
            return isNaN(d.getTime()) ? "No deadline" : d.toLocaleDateString(undefined, {
                month: "short",
                day: "numeric",
                year: "numeric",
                hour: "2-digit",
                minute: "2-digit",
            });
        } catch {
            return "No deadline";
        }
    }, [eventForm.expires_at]);

    return (
        <Card className="shadow-xs border bg-card/60 backdrop-blur-xs">
            <CardHeader
                onClick={() => setIsOpen((prev) => !prev)}
                className="py-3 px-4 cursor-pointer select-none transition-colors hover:bg-muted/30"
            >
                <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex items-center gap-3">
                        <div className="flex items-center gap-2">
                            <Clock className="size-4 text-primary shrink-0" />
                            <CardTitle className="text-xs font-semibold">
                                Registration Limits & Deadline
                            </CardTitle>
                        </div>

                        <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                            <span className="hidden sm:inline text-muted-foreground/50">•</span>
                            <span className="inline-flex items-center gap-1 font-medium text-foreground">
                                <Users className="size-3 text-muted-foreground" />
                                {eventForm.max_applicants === -1 ? "Unlimited Capacity" : `${eventForm.max_applicants} Max`}
                            </span>
                            <span className="hidden sm:inline text-muted-foreground/50">•</span>
                            <span className="inline-flex items-center gap-1 text-muted-foreground">
                                <Calendar className="size-3 text-muted-foreground" />
                                {formattedDeadline}
                            </span>
                        </div>
                    </div>

                    <div className="flex items-center gap-2">
                        {isLocked ? (
                            <Badge
                                variant="secondary"
                                className="text-[10px] gap-1 font-medium text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 border-emerald-500/20"
                            >
                                <Lock className="size-2.5 text-emerald-500" />
                                <span>Locked</span>
                            </Badge>
                        ) : (
                            <Badge variant="outline" className="text-[10px] font-normal text-muted-foreground">
                                Editable
                            </Badge>
                        )}
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-6 text-muted-foreground shrink-0"
                            onClick={(e) => {
                                e.stopPropagation();
                                setIsOpen((prev) => !prev);
                            }}
                        >
                            {isOpen ? <ChevronUp className="size-3.5" /> : <ChevronDown className="size-3.5" />}
                        </Button>
                    </div>
                </div>
            </CardHeader>

            {isOpen && (
                <CardContent className="pt-2 px-4 pb-4 border-t">
                    <CardDescription className="text-xs mb-3 text-muted-foreground">
                        Define maximum attendee capacity and expiration deadline for registrations.
                    </CardDescription>
                    <form onSubmit={handleSave} className="space-y-4">
                        {/* Capacity Section */}
                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <label className="text-xs font-medium text-foreground flex items-center gap-1.5">
                                    <Users className="size-3.5 text-muted-foreground" />
                                    <span>Max Applicants</span>
                                </label>
                                <span className="text-[11px] text-muted-foreground">
                                    Currently registered: <strong>{eventForm.total_applicants}</strong>
                                </span>
                            </div>

                            {isLocked ? (
                                <div className="text-xs font-semibold text-foreground p-2 rounded-md bg-muted/40 border">
                                    {eventForm.max_applicants === -1
                                        ? "Unlimited Attendees"
                                        : `${eventForm.max_applicants} Attendees Max`}
                                </div>
                            ) : (
                                <div className="space-y-2">
                                    <div className="flex items-center gap-2">
                                        <Input
                                            type="number"
                                            min={1}
                                            disabled={isUnlimited || isLocked}
                                            value={isUnlimited ? "" : maxApplicants}
                                            placeholder={isUnlimited ? "Unlimited" : "e.g. 150"}
                                            onChange={(e) =>
                                                setMaxApplicants(Math.max(1, Number(e.target.value)))
                                            }
                                            className="text-xs h-9"
                                        />
                                    </div>
                                    <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer">
                                        <Checkbox
                                            checked={isUnlimited}
                                            disabled={isLocked}
                                            onCheckedChange={(checked) => setIsUnlimited(!!checked)}
                                        />
                                        <span>Unlimited capacity (no applicant cap)</span>
                                    </label>
                                </div>
                            )}
                        </div>

                        {/* Deadline Section */}
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-foreground flex items-center gap-1.5">
                                <Calendar className="size-3.5 text-muted-foreground" />
                                <span>Registration Deadline</span>
                            </label>

                            {isLocked ? (
                                <div className="text-xs font-semibold text-foreground p-2 rounded-md bg-muted/40 border">
                                    {new Date(eventForm.expires_at).toLocaleString()}
                                </div>
                            ) : (
                                <Input
                                    type="datetime-local"
                                    disabled={isLocked}
                                    value={expiresAt}
                                    onChange={(e) => setExpiresAt(e.target.value)}
                                    className="text-xs h-9 font-mono"
                                />
                            )}
                        </div>

                        {!isLocked && (
                            <div className="pt-1 flex justify-end">
                                <Button
                                    type="submit"
                                    disabled={isSaving || !hasChanges}
                                    className="h-8 gap-1.5 text-xs font-semibold"
                                >
                                    <Check className="size-3.5" />
                                    <span>{isSaving ? "Saving..." : "Save Limits"}</span>
                                </Button>
                            </div>
                        )}
                    </form>
                </CardContent>
            )}
        </Card>
    );
}
