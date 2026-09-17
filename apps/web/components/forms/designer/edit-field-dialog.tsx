"use client";

import * as React from "react";
import { Plus, Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
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
import { Switch } from "@/components/ui/switch";
import { FIELD_TYPE_METAS } from "@/components/forms/designer/field-type-config";
import { FileFieldSettings } from "@/components/forms/designer/file-field-settings";
import type { FieldOption, FormField } from "@/schema/forms.types";

interface EditFieldDialogProps {
    field: FormField | null;
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSave: (updatedField: FormField) => void;
}

function EditFieldDialogBody({
    field,
    onOpenChange,
    onSave,
}: {
    field: FormField;
    onOpenChange: (open: boolean) => void;
    onSave: (updatedField: FormField) => void;
}) {
    const [label, setLabel] = React.useState(field.label);
    const [key, setKey] = React.useState(field.key);
    const [placeholder, setPlaceholder] = React.useState(field.placeholder || "");
    const [required, setRequired] = React.useState(field.required);
    const [options, setOptions] = React.useState<FieldOption[]>(
        field.options ? [...field.options] : []
    );
    const [minVal, setMinVal] = React.useState<string>(
        field.validation?.min !== undefined && field.validation?.min !== null
            ? String(field.validation.min)
            : ""
    );
    const [maxVal, setMaxVal] = React.useState<string>(
        field.validation?.max !== undefined && field.validation?.max !== null
            ? String(field.validation.max)
            : ""
    );
    const [minLengthVal, setMinLengthVal] = React.useState<string>(
        field.validation?.min_length !== undefined &&
            field.validation?.min_length !== null
            ? String(field.validation.min_length)
            : ""
    );
    const [maxLengthVal, setMaxLengthVal] = React.useState<string>(
        field.validation?.max_length !== undefined &&
            field.validation?.max_length !== null
            ? String(field.validation.max_length)
            : ""
    );
    const [acceptVal, setAcceptVal] = React.useState<string>(
        field.validation?.accept || ""
    );
    const [maxFileSizeMB, setMaxFileSizeMB] = React.useState<number>(
        field.validation?.max_file_size_mb || 10
    );
    const [allowMultiple, setAllowMultiple] = React.useState<boolean>(
        Boolean(field.validation?.multiple)
    );

    const meta = FIELD_TYPE_METAS[field.type];
    const isOptionsType = field.type === "checkbox" || field.type === "radio";
    const isNumberType = field.type === "number";
    const isLengthType =
        field.type === "text" ||
        field.type === "textarea" ||
        field.type === "number" ||
        field.type === "phone";
    const isFileType = field.type === "file";

    const handleAddOption = () => {
        const nextIdx = options.length + 1;
        setOptions([
            ...options,
            {
                id: `opt_${crypto.randomUUID().slice(0, 6)}`,
                label: `Option ${nextIdx}`,
                value: `option_${nextIdx}`,
            },
        ]);
    };

    const handleOptionChange = (
        index: number,
        keyToUpdate: "label" | "value",
        val: string
    ) => {
        const updated = [...options];
        updated[index] = {
            ...updated[index],
            [keyToUpdate]: val,
        };
        setOptions(updated);
    };

    const handleRemoveOption = (index: number) => {
        setOptions(options.filter((_, idx) => idx !== index));
    };

    const handleSave = () => {
        const updated: FormField = {
            ...field,
            label: label.trim() || field.label,
            key: key.trim() || field.key,
            placeholder: placeholder.trim(),
            required: field.is_system ? true : required,
            options: isOptionsType ? options : [],
            validation: {
                min: isNumberType && minVal !== "" ? parseFloat(minVal) : null,
                max: isNumberType && maxVal !== "" ? parseFloat(maxVal) : null,
                min_length:
                    isLengthType && minLengthVal !== ""
                        ? parseInt(minLengthVal, 10)
                        : null,
                max_length:
                    isLengthType && maxLengthVal !== ""
                        ? parseInt(maxLengthVal, 10)
                        : null,
                accept: isFileType && acceptVal !== "" ? acceptVal.trim() : null,
                max_file_size_mb: isFileType ? maxFileSizeMB : null,
                multiple: isFileType ? allowMultiple : null,
            },
        };

        onSave(updated);
        onOpenChange(false);
    };

    return (
        <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <span>Edit Field Properties</span>
                    <span className="text-xs font-normal text-muted-foreground uppercase px-2 py-0.5 rounded border bg-muted/40">
                        {meta?.label ?? field.type}
                    </span>
                </DialogTitle>
                <DialogDescription>
                    Configure how this field appears and validates data.
                </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-2">
                {/* Field Label */}
                <div className="space-y-1.5">
                    <Label htmlFor="field-label" className="text-xs font-medium">
                        Field Label
                    </Label>
                    <Input
                        id="field-label"
                        value={label}
                        disabled={field.is_system}
                        onChange={(e) => setLabel(e.target.value)}
                        placeholder="e.g., Company Name"
                        className="text-xs h-9"
                    />
                </div>

                {/* Field Key */}
                <div className="space-y-1.5">
                    <Label htmlFor="field-key" className="text-xs font-medium">
                        Field Key
                    </Label>
                    <Input
                        id="field-key"
                        value={key}
                        disabled={true}
                        onChange={(e) =>
                            setKey(
                                e.target.value
                                    .toLowerCase()
                                    .replace(/[^a-z0-9_]/g, "_")
                            )
                        }
                        placeholder="e.g., company_name"
                        className="text-xs h-9 font-mono"
                    />
                    {
                        <p className="text-[11px] text-muted-foreground">
                            System field keys are fixed and cannot be modified.
                        </p>
                    }
                </div>

                {/* Placeholder (if applicable) */}
                {field.type !== "checkbox" &&
                    field.type !== "radio" &&
                    field.type !== "switch" && (
                        <div className="space-y-1.5">
                            <Label
                                htmlFor="field-placeholder"
                                className="text-xs font-medium"
                            >
                                Placeholder Text
                            </Label>
                            <Input
                                id="field-placeholder"
                                value={placeholder}
                                onChange={(e) => setPlaceholder(e.target.value)}
                                placeholder="e.g., Enter your response..."
                                className="text-xs h-9"
                            />
                        </div>
                    )}

                {/* Required Toggle */}
                <div className="flex items-center justify-between rounded-lg border p-3 bg-muted/20">
                    <div className="space-y-0.5">
                        <Label
                            htmlFor="field-required"
                            className="text-xs font-semibold cursor-pointer"
                        >
                            Mandatory / Required Field
                        </Label>
                        <p className="text-[11px] text-muted-foreground">
                            {field.is_system
                                ? "System fields (Name and Email) are always mandatory."
                                : "User must provide an answer to submit the form."}
                        </p>
                    </div>
                    <Switch
                        id="field-required"
                        checked={field.is_system ? true : required}
                        disabled={field.is_system}
                        onCheckedChange={setRequired}
                    />
                </div>

                {/* Options Editor for Checkbox & Radio */}
                {isOptionsType && (
                    <div className="space-y-2 rounded-lg border p-3 bg-muted/10">
                        <div className="flex items-center justify-between">
                            <Label className="text-xs font-semibold">
                                Options List
                            </Label>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={handleAddOption}
                                className="h-7 text-[11px] gap-1 px-2"
                            >
                                <Plus className="size-3" />
                                <span>Add Option</span>
                            </Button>
                        </div>

                        <div className="space-y-2 pt-1">
                            {options.map((opt, idx) => (
                                <div
                                    key={opt.id || idx}
                                    className="flex items-center gap-2"
                                >
                                    <Input
                                        value={opt.label}
                                        onChange={(e) =>
                                            handleOptionChange(
                                                idx,
                                                "label",
                                                e.target.value
                                            )
                                        }
                                        placeholder="Option label"
                                        className="text-xs h-8 flex-1"
                                    />
                                    <Input
                                        value={opt.value}
                                        disabled={true}
                                        onChange={(e) =>
                                            handleOptionChange(
                                                idx,
                                                "value",
                                                e.target.value
                                            )
                                        }
                                        placeholder="value"
                                        className="text-xs h-8 w-28 font-mono"
                                    />
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon-sm"
                                        disabled={options.length <= 1}
                                        onClick={() => handleRemoveOption(idx)}
                                        className="text-destructive hover:bg-destructive/10"
                                    >
                                        <Trash2 className="size-3.5" />
                                    </Button>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {/* Numeric Min/Max Value Limits */}
                {isNumberType && (
                    <div className="grid grid-cols-2 gap-3 rounded-lg border p-3 bg-muted/10">
                        <div className="space-y-1">
                            <Label className="text-xs">Minimum Value</Label>
                            <Input
                                type="number"
                                value={minVal}
                                onChange={(e) => setMinVal(e.target.value)}
                                placeholder="No limit"
                                className="text-xs h-8"
                            />
                        </div>
                        <div className="space-y-1">
                            <Label className="text-xs">Maximum Value</Label>
                            <Input
                                type="number"
                                value={maxVal}
                                onChange={(e) => setMaxVal(e.target.value)}
                                placeholder="No limit"
                                className="text-xs h-8"
                            />
                        </div>
                    </div>
                )}

                {/* Min / Max Length for Text, Textarea, Number */}
                {isLengthType && (
                    <div className="grid grid-cols-2 gap-3 rounded-lg border p-3 bg-muted/10">
                        <div className="space-y-1">
                            <Label className="text-xs">
                                {field.type === "number"
                                    ? "Min Digits / Length"
                                    : "Minimum Characters"}
                            </Label>
                            <Input
                                type="number"
                                min={0}
                                value={minLengthVal}
                                onChange={(e) => setMinLengthVal(e.target.value)}
                                placeholder="e.g., 2"
                                className="text-xs h-8"
                            />
                        </div>
                        <div className="space-y-1">
                            <Label className="text-xs">
                                {field.type === "number"
                                    ? "Max Digits / Length"
                                    : "Maximum Characters"}
                            </Label>
                            <Input
                                type="number"
                                min={0}
                                value={maxLengthVal}
                                onChange={(e) => setMaxLengthVal(e.target.value)}
                                placeholder="e.g., 100"
                                className="text-xs h-8"
                            />
                        </div>
                    </div>
                )}

                {/* File Upload Accept & Max File Size Range */}
                {isFileType && (
                    <FileFieldSettings
                        acceptValue={acceptVal}
                        onAcceptChange={setAcceptVal}
                        maxFileSizeMB={maxFileSizeMB}
                        onMaxFileSizeChange={setMaxFileSizeMB}
                        allowMultiple={allowMultiple}
                        onAllowMultipleChange={setAllowMultiple}
                    />
                )}
            </div>

            <DialogFooter className="pt-2">
                <Button
                    type="button"
                    variant="outline"
                    onClick={() => onOpenChange(false)}
                >
                    Cancel
                </Button>
                <Button type="button" onClick={handleSave}>
                    Save Changes
                </Button>
            </DialogFooter>
        </DialogContent>
    );
}

export function EditFieldDialog({
    field,
    open,
    onOpenChange,
    onSave,
}: EditFieldDialogProps) {
    if (!field) return null;

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <EditFieldDialogBody
                key={field.id}
                field={field}
                onOpenChange={onOpenChange}
                onSave={onSave}
            />
        </Dialog>
    );
}
