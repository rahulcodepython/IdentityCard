"use client";

import * as React from "react";
import { Plus } from "lucide-react";

import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
import { Input } from "../ui/input";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "../ui/select";
import type {
    ApplicantFilter,
    FormFieldOption,
    FormFieldSummary,
} from "../../schema/applicants.types";

interface ApplicantFilterFormProps {
    fields: FormFieldSummary[];
    onAddFilter: (filter: ApplicantFilter) => void;
}

interface ParsedOption {
    label: string;
    value: string;
}

function parseOptions(
    rawOptions: (string | FormFieldOption)[] | null | undefined
): ParsedOption[] {
    if (!rawOptions || !Array.isArray(rawOptions)) return [];
    return rawOptions.map((opt) => {
        if (typeof opt === "string") {
            return { label: opt, value: opt };
        }
        return {
            label: opt.label || opt.value,
            value: opt.value,
        };
    });
}

export function ApplicantFilterForm({
    fields,
    onAddFilter,
}: ApplicantFilterFormProps) {
    const availableFields = React.useMemo(() => {
        const list: {
            key: string;
            label: string;
            type: string;
            options: ParsedOption[];
        }[] = [
            { key: "name", label: "Full Name", type: "text", options: [] },
            { key: "email", label: "Email Address", type: "email", options: [] },
            { key: "phone", label: "Mobile Number", type: "phone", options: [] },
            { key: "user_id", label: "Applicant ID", type: "text", options: [] },
        ];

        for (const f of fields) {
            if (
                f.key !== "name" &&
                f.key !== "email" &&
                f.key !== "phone" &&
                f.key !== "mobile" &&
                f.key !== "user_id"
            ) {
                list.push({
                    key: f.key,
                    label: f.label || f.key,
                    type: f.type,
                    options: parseOptions(f.options),
                });
            }
        }
        return list;
    }, [fields]);

    const [selectedFieldKey, setSelectedFieldKey] = React.useState<string>(
        availableFields[0]?.key || "name"
    );

    const currentField = React.useMemo(() => {
        return (
            availableFields.find((f) => f.key === selectedFieldKey) ||
            availableFields[0]
        );
    }, [availableFields, selectedFieldKey]);

    // Value states for each specialized field type
    const [selectedCheckboxValues, setSelectedCheckboxValues] = React.useState<
        string[]
    >([]);
    const [selectedRadioValue, setSelectedRadioValue] = React.useState<string>("");
    const [switchValue, setSwitchValue] = React.useState<"true" | "false">("true");
    const [dateMode, setDateMode] = React.useState<"exact" | "range">("exact");
    const [exactDate, setExactDate] = React.useState<string>("");
    const [startDate, setStartDate] = React.useState<string>("");
    const [endDate, setEndDate] = React.useState<string>("");
    const [textValue, setTextValue] = React.useState<string>("");

    // Reset inputs when switching fields
    const prevFieldKeyRef = React.useRef(selectedFieldKey);
    React.useEffect(() => {
        if (prevFieldKeyRef.current !== selectedFieldKey) {
            prevFieldKeyRef.current = selectedFieldKey;
            setSelectedCheckboxValues([]);
            setSelectedRadioValue(currentField?.options[0]?.value || "");
            setSwitchValue("true");
            setExactDate("");
            setStartDate("");
            setEndDate("");
            setTextValue("");
        }
    }, [selectedFieldKey, currentField]);

    const handleApply = (e: React.FormEvent) => {
        e.preventDefault();
        const fieldKey = currentField.key;
        const fieldType = currentField.type;

        // 1. Checkbox with options: filter with one or many selected options
        if (fieldType === "checkbox" && currentField.options.length > 0) {
            if (selectedCheckboxValues.length === 0) return;
            for (const val of selectedCheckboxValues) {
                onAddFilter({
                    field: fieldKey,
                    op: "contains",
                    value: val,
                });
            }
            setSelectedCheckboxValues([]);
            return;
        }

        // 2. Radio: filter with one selected option
        if (fieldType === "radio" || (fieldType === "select" && currentField.options.length > 0)) {
            if (!selectedRadioValue) return;
            onAddFilter({
                field: fieldKey,
                op: "eq",
                value: selectedRadioValue,
            });
            return;
        }

        // 3. Switch or Boolean single checkbox: true / false
        if (
            fieldType === "switch" ||
            (fieldType === "checkbox" && currentField.options.length === 0)
        ) {
            onAddFilter({
                field: fieldKey,
                op: "eq",
                value: switchValue,
            });
            return;
        }

        // 4. Date / Time / Week / Month: exact date or date range
        if (
            fieldType === "date" ||
            fieldType === "time" ||
            fieldType === "week" ||
            fieldType === "month"
        ) {
            if (dateMode === "exact") {
                if (!exactDate) return;
                onAddFilter({
                    field: fieldKey,
                    op: "eq",
                    value: exactDate,
                });
                setExactDate("");
            } else {
                if (!startDate && !endDate) return;
                if (startDate) {
                    onAddFilter({
                        field: fieldKey,
                        op: "gte",
                        value: startDate,
                    });
                }
                if (endDate) {
                    onAddFilter({
                        field: fieldKey,
                        op: "lte",
                        value: endDate,
                    });
                }
                setStartDate("");
                setEndDate("");
            }
            return;
        }

        // 5. Default: Text, Textarea, Email, Number, URL, Applicant ID -> uses contains
        if (!textValue.trim()) return;
        onAddFilter({
            field: fieldKey,
            op: "contains",
            value: textValue.trim(),
        });
        setTextValue("");
    };

    const isApplyDisabled = React.useMemo(() => {
        const fieldType = currentField.type;
        if (fieldType === "checkbox" && currentField.options.length > 0) {
            return selectedCheckboxValues.length === 0;
        }
        if (fieldType === "radio" || (fieldType === "select" && currentField.options.length > 0)) {
            return !selectedRadioValue;
        }
        if (
            fieldType === "date" ||
            fieldType === "time" ||
            fieldType === "week" ||
            fieldType === "month"
        ) {
            return dateMode === "exact" ? !exactDate : !startDate && !endDate;
        }
        if (fieldType === "switch" || (fieldType === "checkbox" && currentField.options.length === 0)) {
            return false;
        }
        return !textValue.trim();
    }, [
        currentField,
        selectedCheckboxValues,
        selectedRadioValue,
        dateMode,
        exactDate,
        startDate,
        endDate,
        textValue,
    ]);

    return (
        <form
            onSubmit={handleApply}
            className="flex flex-col gap-3 rounded-lg border bg-card p-3.5 shadow-xs"
        >
            <div className="flex flex-wrap items-end gap-3">
                {/* 1. Attribute / Field Selector */}
                <div className="flex flex-col gap-1.5 min-w-48">
                    <label className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                        Attribute
                    </label>
                    <Select
                        value={selectedFieldKey}
                        onValueChange={(val) => {
                            if (val) setSelectedFieldKey(val);
                        }}
                    >
                        <SelectTrigger className="h-10 text-sm">
                            <SelectValue placeholder="Select attribute" />
                        </SelectTrigger>
                        <SelectContent>
                            {availableFields.map((f) => (
                                <SelectItem key={f.key} value={f.key}>
                                    {f.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>

                {/* 2. Specialized Inputs (NO OPERATOR DROPDOWN) */}
                <div className="flex flex-col gap-1.5 flex-1 min-w-56">
                    <label className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                        Filter Value
                    </label>

                    {/* A. Checkbox with options: Multi-select choices */}
                    {currentField.type === "checkbox" &&
                        currentField.options.length > 0 && (
                            <div className="flex flex-wrap items-center gap-2 pt-0.5">
                                {currentField.options.map((opt) => {
                                    const isChecked =
                                        selectedCheckboxValues.includes(opt.value);
                                    return (
                                        <label
                                            key={opt.value}
                                            className="flex items-center gap-1.5 text-xs text-foreground cursor-pointer rounded-md border px-2.5 py-1 bg-background hover:bg-muted/40 transition-colors"
                                        >
                                            <Checkbox
                                                checked={isChecked}
                                                onCheckedChange={(checked) => {
                                                    if (checked) {
                                                        setSelectedCheckboxValues(
                                                            (prev) => [
                                                                ...prev,
                                                                opt.value,
                                                            ]
                                                        );
                                                    } else {
                                                        setSelectedCheckboxValues(
                                                            (prev) =>
                                                                prev.filter(
                                                                    (v) =>
                                                                        v !==
                                                                        opt.value
                                                                )
                                                        );
                                                    }
                                                }}
                                            />
                                            <span>{opt.label}</span>
                                        </label>
                                    );
                                })}
                            </div>
                        )}

                    {/* B. Radio or Select with options: Single choice */}
                    {(currentField.type === "radio" ||
                        (currentField.type === "select" &&
                            currentField.options.length > 0)) && (
                            <div className="flex flex-wrap items-center gap-1.5 pt-0.5">
                                {currentField.options.map((opt) => {
                                    const isSelected = selectedRadioValue === opt.value;
                                    return (
                                        <Badge
                                            key={opt.value}
                                            variant={isSelected ? "default" : "outline"}
                                            onClick={() => setSelectedRadioValue(opt.value)}
                                            className="cursor-pointer text-xs font-normal px-2.5 py-1 transition-colors"
                                        >
                                            {opt.label}
                                        </Badge>
                                    );
                                })}
                            </div>
                        )}

                    {/* C. Switch or Single Checkbox: True / False */}
                    {(currentField.type === "switch" ||
                        (currentField.type === "checkbox" &&
                            currentField.options.length === 0)) && (
                            <div className="flex items-center gap-2 pt-0.5">
                                <Badge
                                    variant={switchValue === "true" ? "default" : "outline"}
                                    onClick={() => setSwitchValue("true")}
                                    className="cursor-pointer text-xs px-3 py-1"
                                >
                                    Yes (True)
                                </Badge>
                                <Badge
                                    variant={switchValue === "false" ? "default" : "outline"}
                                    onClick={() => setSwitchValue("false")}
                                    className="cursor-pointer text-xs px-3 py-1"
                                >
                                    No (False)
                                </Badge>
                            </div>
                        )}

                    {/* D. Date / Time / Week / Month: Exact Date or Range */}
                    {(currentField.type === "date" ||
                        currentField.type === "time" ||
                        currentField.type === "week" ||
                        currentField.type === "month") && (
                            <div className="flex flex-wrap items-center gap-2">
                                <div className="flex rounded-md border p-0.5 bg-muted/40 text-[11px]">
                                    <button
                                        type="button"
                                        onClick={() => setDateMode("exact")}
                                        className={`px-2 py-0.5 rounded font-medium transition-colors ${dateMode === "exact"
                                                ? "bg-background shadow-xs text-foreground"
                                                : "text-muted-foreground hover:text-foreground"
                                            }`}
                                    >
                                        Exact
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setDateMode("range")}
                                        className={`px-2 py-0.5 rounded font-medium transition-colors ${dateMode === "range"
                                                ? "bg-background shadow-xs text-foreground"
                                                : "text-muted-foreground hover:text-foreground"
                                            }`}
                                    >
                                        Range
                                    </button>
                                </div>

                                {dateMode === "exact" ? (
                                    <Input
                                        type={
                                            currentField.type === "time"
                                                ? "time"
                                                : currentField.type === "month"
                                                    ? "month"
                                                    : currentField.type === "week"
                                                        ? "week"
                                                        : "date"
                                        }
                                        value={exactDate}
                                        onChange={(e) => setExactDate(e.target.value)}
                                        className="h-10 text-sm w-48"
                                    />
                                ) : (
                                    <div className="flex items-center gap-2">
                                        <Input
                                            type={
                                                currentField.type === "time"
                                                    ? "time"
                                                    : currentField.type === "month"
                                                        ? "month"
                                                        : currentField.type === "week"
                                                            ? "week"
                                                            : "date"
                                            }
                                            value={startDate}
                                            onChange={(e) => setStartDate(e.target.value)}
                                            placeholder="From"
                                            className="h-10 text-sm w-40"
                                        />
                                        <span className="text-sm text-muted-foreground">
                                            to
                                        </span>
                                        <Input
                                            type={
                                                currentField.type === "time"
                                                    ? "time"
                                                    : currentField.type === "month"
                                                        ? "month"
                                                        : currentField.type === "week"
                                                            ? "week"
                                                            : "date"
                                            }
                                            value={endDate}
                                            onChange={(e) => setEndDate(e.target.value)}
                                            placeholder="To"
                                            className="h-10 text-sm w-40"
                                        />
                                    </div>
                                )}
                            </div>
                        )}

                    {/* E. Text, Textarea, Email, Number, URL, ID: default contains */}
                    {currentField.type !== "checkbox" &&
                        currentField.type !== "radio" &&
                        currentField.type !== "select" &&
                        currentField.type !== "switch" &&
                        currentField.type !== "date" &&
                        currentField.type !== "time" &&
                        currentField.type !== "week" &&
                        currentField.type !== "month" && (
                            <Input
                                type="text"
                                placeholder={`Filter by ${currentField.label.toLowerCase()}...`}
                                value={textValue}
                                onChange={(e) => setTextValue(e.target.value)}
                                className="h-10 text-sm"
                            />
                        )}
                </div>

                {/* 3. Add Filter Action Button */}
                <div className="flex items-end">
                    <Button
                        type="submit"
                        disabled={isApplyDisabled}
                        className="h-10 gap-2 px-4 text-sm font-semibold"
                    >
                        <Plus className="size-4" />
                        <span>Add Filter</span>
                    </Button>
                </div>
            </div>
        </form>
    );
}
