"use client";

import * as React from "react";
import { RotateCcw, X } from "lucide-react";

import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import type { ApplicantFilter, FormFieldSummary } from "../../schema/applicants.types";
import { getOptionDisplayLabel } from "./applicants-utils";

interface ApplicantFilterBadgesProps {
    filters: ApplicantFilter[];
    fields: FormFieldSummary[];
    onRemoveFilter: (index: number) => void;
    onClearFilters: () => void;
}

function formatFilterBadgeText(
    flt: ApplicantFilter,
    fields: FormFieldSummary[]
): { fieldLabel: string; valueText: string } {
    let fieldLabel = flt.field;
    if (flt.field === "name") fieldLabel = "Full Name";
    else if (flt.field === "email") fieldLabel = "Email";
    else if (flt.field === "user_id") fieldLabel = "Applicant ID";

    const fieldDef = fields.find((f) => f.key === flt.field);
    if (fieldDef?.label) {
        fieldLabel = fieldDef.label;
    }

    const fieldType = fieldDef?.type || (flt.field === "email" ? "email" : "text");

    // 1. Checkbox or Switch (Boolean true/false)
    if (
        fieldType === "switch" ||
        (fieldType === "checkbox" && (!fieldDef?.options || fieldDef.options.length === 0))
    ) {
        const isYes = flt.value === "true" || flt.value === "1";
        return { fieldLabel, valueText: isYes ? "Yes" : "No" };
    }

    // 2. Radio, Select, or Multi-Option Checkbox: use option label instead of raw value
    if (
        fieldType === "radio" ||
        fieldType === "select" ||
        (fieldType === "checkbox" && fieldDef?.options && fieldDef.options.length > 0)
    ) {
        const display = getOptionDisplayLabel(fieldDef?.options, flt.value);
        return { fieldLabel, valueText: display };
    }

    // 3. Date / Time / Week / Month
    if (
        fieldType === "date" ||
        fieldType === "time" ||
        fieldType === "week" ||
        fieldType === "month"
    ) {
        if (flt.op === "gte") {
            return { fieldLabel, valueText: `From ${flt.value}` };
        }
        if (flt.op === "lte") {
            return { fieldLabel, valueText: `To ${flt.value}` };
        }
        return { fieldLabel, valueText: flt.value };
    }

    // 4. Default: Text, Textarea, Email, Number, URL, Applicant ID
    return { fieldLabel, valueText: `"${flt.value}"` };
}

export function ApplicantFilterBadges({
    filters,
    fields,
    onRemoveFilter,
    onClearFilters,
}: ApplicantFilterBadgesProps) {
    if (filters.length === 0) return null;

    return (
        <div className="flex flex-wrap items-center gap-1.5 pt-1">
            <span className="text-[11px] text-muted-foreground mr-1">
                Active filters:
            </span>
            {filters.map((flt, idx) => {
                const { fieldLabel, valueText } = formatFilterBadgeText(
                    flt,
                    fields
                );

                return (
                    <Badge
                        key={`${flt.field}-${flt.op}-${flt.value}-${idx}`}
                        variant="secondary"
                        className="gap-1.5 pl-2.5 pr-1 py-0.5 text-[11px] font-normal border shadow-2xs"
                    >
                        <span className="font-semibold text-foreground">
                            {fieldLabel}:
                        </span>
                        <span className="text-foreground font-medium">
                            {valueText}
                        </span>
                        <button
                            type="button"
                            onClick={() => onRemoveFilter(idx)}
                            className="rounded-full p-0.5 hover:bg-muted text-muted-foreground hover:text-foreground transition-colors ml-0.5"
                            title="Remove filter"
                        >
                            <X className="size-3" />
                        </button>
                    </Badge>
                );
            })}

            <Button
                type="button"
                variant="ghost"
                onClick={onClearFilters}
                className="text-muted-foreground hover:text-destructive gap-1.5"
            >
                <RotateCcw className="size-3.5" />
                <span>Clear all</span>
            </Button>
        </div>
    );
}
