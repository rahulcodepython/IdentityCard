"use client";

import * as React from "react";
import {
    AlertTriangle,
    Calendar,
    Loader2,
    Lock,
    QrCode,
    Save,
    SlidersHorizontal,
    Trash2,
    Users,
} from "lucide-react";

import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { EventFormDetails } from "@/schema/eventform.types";

interface EventFormHeaderProps {
    eventForm: EventFormDetails;
    eventName: string;
    onLockClick: () => void;
    isLocking: boolean;
    onDeleteForm: () => void;
    isDeleting: boolean;
    onOpenConfig: () => void;
    onOpenShare: () => void;
    hasUnsavedChanges: boolean;
    onSaveChanges: () => void;
    isSaving: boolean;
}

export function EventFormHeader({
    eventForm,
    eventName,
    onLockClick,
    isLocking,
    onDeleteForm,
    isDeleting,
    onOpenConfig,
    onOpenShare,
    hasUnsavedChanges,
    onSaveChanges,
    isSaving,
}: EventFormHeaderProps) {
    const isLocked = eventForm.is_locked;
    const canDelete = eventForm.can_delete && eventForm.total_applicants === 0;

    const formattedDeadline = React.useMemo(() => {
        try {
            const d = new Date(eventForm.expires_at);
            return isNaN(d.getTime())
                ? "No deadline"
                : d.toLocaleDateString(undefined, {
                    month: "short",
                    day: "numeric",
                    year: "numeric",
                });
        } catch {
            return "No deadline";
        }
    }, [eventForm.expires_at]);

    return (
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between px-1 shrink-0">
            {/* Title & Status */}
            <div>
                <div className="flex items-center gap-2">
                    <h1 className="text-lg font-bold text-foreground">
                        {eventForm.name || "Event Registration Form"}
                    </h1>
                    {isLocked ? (
                        <Badge
                            variant="default"
                            className="bg-emerald-600 hover:bg-emerald-600 text-[10px] gap-1 font-semibold"
                        >
                            <Lock className="size-3" />
                            <span>Locked & Live</span>
                        </Badge>
                    ) : (
                        <Badge
                            variant="secondary"
                            className="text-[10px] gap-1 font-normal text-muted-foreground"
                        >
                            <span>Draft (Unlocked)</span>
                        </Badge>
                    )}
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground mt-0.5">
                    <span>Registration for {eventName || "this event"}.</span>
                    <span className="text-muted-foreground/40 hidden sm:inline">•</span>
                    <span className="inline-flex items-center gap-1 font-medium text-foreground">
                        <Users className="size-3 text-muted-foreground" />
                        {eventForm.max_applicants === -1
                            ? "Unlimited"
                            : `${eventForm.max_applicants} Max`}
                    </span>
                    <span className="text-muted-foreground/40 hidden sm:inline">•</span>
                    <span className="inline-flex items-center gap-1 text-muted-foreground">
                        <Calendar className="size-3 text-muted-foreground" />
                        {formattedDeadline}
                    </span>
                </div>
            </div>

            {/* Top-right Actions */}
            <div className="flex items-center gap-2 flex-wrap">
                {/* 1. Registration Config Dialog Button */}
                <Button
                    type="button"
                    variant="outline"
                    onClick={onOpenConfig}
                    className="gap-1.5 text-xs font-medium"
                >
                    <SlidersHorizontal className="size-3.5 text-primary" />
                    <span>Registration Config</span>
                </Button>

                {/* 2. Unlocked Actions */}
                {!isLocked ? (
                    <>
                        {/* Delete Form Dialog */}
                        <AlertDialog>
                            <AlertDialogTrigger
                                render={
                                    <Button
                                        type="button"
                                        variant="outline"
                                        disabled={!canDelete || isDeleting}
                                        className="gap-1.5 text-xs font-medium text-destructive hover:text-destructive hover:bg-destructive/10"
                                        title={
                                            canDelete
                                                ? "Delete this form"
                                                : "Cannot delete form with registered applicants"
                                        }
                                    />
                                }
                            >
                                {isDeleting ? (
                                    <Loader2 className="size-3.5 animate-spin" />
                                ) : (
                                    <Trash2 className="size-3.5" />
                                )}
                                <span>Delete Form</span>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                                <AlertDialogHeader>
                                    <AlertDialogTitle className="text-sm font-semibold flex items-center gap-2">
                                        <AlertTriangle className="size-4 text-destructive" />
                                        <span>Delete Event Form?</span>
                                    </AlertDialogTitle>
                                    <AlertDialogDescription className="text-xs">
                                        This will permanently remove the registration form from this event. You will be able to set up a new form from a template or build from scratch.
                                    </AlertDialogDescription>
                                </AlertDialogHeader>
                                <AlertDialogFooter>
                                    <AlertDialogCancel className="text-xs">
                                        Cancel
                                    </AlertDialogCancel>
                                    <AlertDialogAction
                                        onClick={onDeleteForm}
                                        className="text-xs bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                    >
                                        Yes, Delete Form
                                    </AlertDialogAction>
                                </AlertDialogFooter>
                            </AlertDialogContent>
                        </AlertDialog>

                        {/* Save Changes Button */}
                        {hasUnsavedChanges && (
                            <Button
                                type="button"
                                variant="secondary"
                                disabled={isSaving}
                                onClick={onSaveChanges}
                                className="gap-1.5 text-xs font-semibold"
                            >
                                {isSaving ? (
                                    <Loader2 className="size-3.5 animate-spin" />
                                ) : (
                                    <Save className="size-3.5" />
                                )}
                                <span>Save</span>
                            </Button>
                        )}

                        {/* Lock Form Trigger Button */}
                        <Button
                            type="button"
                            variant="default"
                            disabled={isLocking || hasUnsavedChanges}
                            onClick={onLockClick}
                            className="gap-1.5 text-xs font-semibold"
                        >
                            {isLocking ? (
                                <Loader2 className="size-3.5 animate-spin" />
                            ) : (
                                <Lock className="size-3.5" />
                            )}
                            <span>Lock Form</span>
                        </Button>
                    </>
                ) : (
                    /* 3. Locked Actions: QR Code & Share */
                    <Button
                        type="button"
                        variant="default"
                        onClick={onOpenShare}
                        className="gap-1.5 text-xs font-semibold"
                    >
                        <QrCode className="size-3.5" />
                        <span>QR Code & Share</span>
                    </Button>
                )}
            </div>
        </div>
    );
}
