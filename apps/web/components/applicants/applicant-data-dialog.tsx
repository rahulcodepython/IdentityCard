"use client";

import * as React from "react";

import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import type { ApplicantItem, FormFieldSummary, FormFieldValue } from "@/schema/applicants.types";
import { formatFieldValue } from "./applicants-utils";

interface ApplicantDataDialogProps {
    applicant: ApplicantItem;
    formFields?: FormFieldSummary[];
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function ApplicantDataDialog({
    applicant,
    formFields,
    open,
    onOpenChange,
}: ApplicantDataDialogProps) {
    const entries = Object.entries(applicant.data || {});

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle className="text-base font-semibold">
                        Applicant Submission Details
                    </DialogTitle>
                    <DialogDescription className="text-xs">
                        Registered data for{" "}
                        <span className="font-semibold text-foreground">
                            {applicant.name}
                        </span>{" "}
                        ({applicant.user_id}).
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-3 pt-2 max-h-[60vh] overflow-y-auto">
                    <div className="rounded-lg border p-3 bg-muted/20 space-y-1 text-xs">
                        <div className="flex justify-between">
                            <span className="text-muted-foreground">Full Name:</span>
                            <span className="font-semibold text-foreground">
                                {applicant.name}
                            </span>
                        </div>
                        <div className="flex justify-between">
                            <span className="text-muted-foreground">Email:</span>
                            <span className="font-mono text-foreground">
                                {applicant.email}
                            </span>
                        </div>
                        <div className="flex justify-between">
                            <span className="text-muted-foreground">Applied At:</span>
                            <span className="text-foreground">
                                {new Date(applicant.created_at).toLocaleString()}
                            </span>
                        </div>
                    </div>

                    <div className="space-y-2">
                        <span className="text-xs font-semibold text-foreground block">
                            Form Fields
                        </span>
                        {entries.length === 0 ? (
                            <p className="text-xs text-muted-foreground italic">
                                No additional custom field responses submitted.
                            </p>
                        ) : (
                            <div className="divide-y rounded-lg border bg-card">
                                {entries.map(([key, val]) => {
                                    const fieldDef = formFields?.find(
                                        (f) => f.key === key
                                    );
                                    const displayLabel =
                                        fieldDef?.label ||
                                        key.replace(/_/g, " ");
                                    const displayVal = fieldDef
                                        ? formatFieldValue(
                                              fieldDef,
                                              val as FormFieldValue
                                          )
                                        : typeof val === "boolean"
                                          ? val
                                              ? "Yes"
                                              : "No"
                                          : Array.isArray(val)
                                            ? val.join(", ")
                                            : String(val ?? "-");

                                    return (
                                        <div
                                            key={key}
                                            className="flex items-start justify-between gap-4 p-2.5 text-xs"
                                        >
                                            <span className="font-mono text-[11px] text-muted-foreground capitalize">
                                                {displayLabel}:
                                            </span>
                                            <span className="font-medium text-foreground text-right break-all">
                                                {displayVal}
                                            </span>
                                        </div>
                                    );
                                })}
                            </div>
                        )}
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}
