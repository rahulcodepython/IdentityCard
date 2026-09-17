"use client";

import * as React from "react";
import { type Control } from "react-hook-form";
import { UploadCloud } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
import {
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PhoneInput } from "@/components/ui/phone-input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import type { FormField as FormFieldType } from "@/schema/forms.types";
import type { PublicApplyFormValues } from "@/schema/publicapply.types";

interface ApplyFieldInputProps {
    control: Control<PublicApplyFormValues>;
    field: FormFieldType;
}

export function ApplyFieldInput({ control, field }: ApplyFieldInputProps) {
    const fieldKey = field.key;

    return (
        <FormField
            control={control}
            name={`data.${fieldKey}`}
            render={({ field: formField }) => (
                <FormItem className="space-y-1.5">
                    <FormLabel className="text-xs font-medium">
                        {field.label}
                        {field.required && (
                            <span className="text-destructive font-bold ml-1">*</span>
                        )}
                    </FormLabel>

                    <FormControl>
                        <div>
                            {/* Textarea */}
                            {field.type === "textarea" && (
                                <Textarea
                                    placeholder={
                                        field.placeholder ||
                                        `Enter ${field.label.toLowerCase()}`
                                    }
                                    value={
                                        typeof formField.value === "string"
                                            ? formField.value
                                            : ""
                                    }
                                    onChange={(e) =>
                                        formField.onChange(e.target.value)
                                    }
                                    minLength={
                                        field.validation?.min_length ?? undefined
                                    }
                                    maxLength={
                                        field.validation?.max_length ?? undefined
                                    }
                                    className="text-xs min-h-20"
                                />
                            )}

                            {/* Text, Email, URL, Date, Time, Month, Week */}
                            {(field.type === "text" ||
                                field.type === "email" ||
                                field.type === "url" ||
                                field.type === "date" ||
                                field.type === "time" ||
                                field.type === "month" ||
                                field.type === "week") && (
                                <Input
                                    type={field.type}
                                    placeholder={
                                        field.placeholder ||
                                        `Enter ${field.label.toLowerCase()}`
                                    }
                                    value={
                                        typeof formField.value === "string"
                                            ? formField.value
                                            : ""
                                    }
                                    onChange={(e) =>
                                        formField.onChange(e.target.value)
                                    }
                                    minLength={
                                        field.validation?.min_length ?? undefined
                                    }
                                    maxLength={
                                        field.validation?.max_length ?? undefined
                                    }
                                    className="text-xs h-9"
                                />
                            )}

                            {/* Mobile Number / Phone */}
                            {field.type === "phone" && (
                                <PhoneInput
                                    placeholder={
                                        field.placeholder ||
                                        "10-digit mobile number"
                                    }
                                    value={
                                        typeof formField.value === "string"
                                            ? formField.value
                                            : ""
                                    }
                                    onChange={(val) => formField.onChange(val)}
                                />
                            )}

                            {/* Number */}
                            {field.type === "number" && (
                                <Input
                                    type="number"
                                    placeholder={
                                        field.placeholder ||
                                        `Enter ${field.label.toLowerCase()}`
                                    }
                                    value={
                                        typeof formField.value === "number" ||
                                        typeof formField.value === "string"
                                            ? formField.value
                                            : ""
                                    }
                                    min={field.validation?.min ?? undefined}
                                    max={field.validation?.max ?? undefined}
                                    onChange={(e) =>
                                        formField.onChange(
                                            e.target.value === ""
                                                ? ""
                                                : Number(e.target.value)
                                        )
                                    }
                                    className="text-xs h-9"
                                />
                            )}

                            {/* Checkbox Group */}
                            {field.type === "checkbox" && (
                                <div className="space-y-2 pt-1">
                                    {(field.options || []).map((opt) => {
                                        const currentArr = Array.isArray(
                                            formField.value
                                        )
                                            ? (formField.value as string[])
                                            : [];
                                        const isChecked = currentArr.includes(
                                            opt.value
                                        );

                                        return (
                                            <label
                                                key={opt.id}
                                                className="flex items-center gap-2.5 text-xs text-foreground cursor-pointer"
                                            >
                                                <Checkbox
                                                    checked={isChecked}
                                                    onCheckedChange={(checked) => {
                                                        if (checked) {
                                                            formField.onChange([
                                                                ...currentArr,
                                                                opt.value,
                                                            ]);
                                                        } else {
                                                            formField.onChange(
                                                                currentArr.filter(
                                                                    (v: string) =>
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

                            {/* Radio Group */}
                            {field.type === "radio" && (
                                <div className="space-y-2 pt-1">
                                    {(field.options || []).map((opt) => (
                                        <label
                                            key={opt.id}
                                            className="flex items-center gap-2.5 text-xs text-foreground cursor-pointer"
                                        >
                                            <input
                                                type="radio"
                                                name={`radio-${fieldKey}`}
                                                value={opt.value}
                                                checked={
                                                    formField.value === opt.value
                                                }
                                                onChange={() =>
                                                    formField.onChange(opt.value)
                                                }
                                                className="size-3.5 text-primary border-input focus:ring-ring"
                                            />
                                            <span>{opt.label}</span>
                                        </label>
                                    ))}
                                </div>
                            )}

                            {/* Switch Single Toggle */}
                            {field.type === "switch" && (
                                <div className="flex items-center justify-between p-2.5 rounded-lg border bg-muted/10">
                                    <span className="text-xs text-muted-foreground">
                                        {field.placeholder || "Enable option"}
                                    </span>
                                    <Switch
                                        checked={Boolean(formField.value)}
                                        onCheckedChange={(checked) =>
                                            formField.onChange(checked)
                                        }
                                    />
                                </div>
                            )}

                            {/* File Upload Field */}
                            {field.type === "file" && (
                                <div className="space-y-2">
                                    <label
                                        htmlFor={`field-${field.id}`}
                                        className="flex flex-col items-center justify-center p-4 border border-dashed rounded-lg bg-muted/20 hover:bg-muted/40 cursor-pointer transition-colors text-center"
                                    >
                                        <UploadCloud className="size-6 text-muted-foreground mb-1" />
                                        <span className="text-xs font-medium text-foreground">
                                            {formField.value
                                                ? typeof formField.value === "string"
                                                    ? formField.value
                                                    : "File selected"
                                                : "Click to select file"}
                                        </span>
                                        <span className="text-[10px] text-muted-foreground mt-0.5">
                                            {field.validation?.accept ||
                                                "Any document format"}{" "}
                                            (Max{" "}
                                            {field.validation?.max_file_size_mb ||
                                                10}
                                            MB)
                                        </span>
                                    </label>
                                    <input
                                        id={`field-${field.id}`}
                                        type="file"
                                        accept={
                                            field.validation?.accept ||
                                            undefined
                                        }
                                        className="hidden"
                                        onChange={(e) => {
                                            const file = e.target.files?.[0];
                                            if (file) {
                                                formField.onChange(file.name);
                                            }
                                        }}
                                    />
                                </div>
                            )}
                        </div>
                    </FormControl>

                    <FormMessage className="text-[11px]" />
                </FormItem>
            )}
        />
    );
}
