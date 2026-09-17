"use client";

import * as React from "react";
import {
    Calendar,
    Clock,
    MapPin,
    ShieldCheck,
    UploadCloud,
    UserCheck,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PhoneInput } from "@/components/ui/phone-input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import type { Form, FormField } from "@/schema/forms.types";

interface FormPreviewPanelProps {
    form?: { name?: string } | Form;
    fields: FormField[];
}

export function FormPreviewPanel({ form, fields }: FormPreviewPanelProps) {
    // Local state for preview interactions
    const [formValues, setFormValues] = React.useState<Record<string, unknown>>({});

    const handleTextChange = (key: string, val: string) => {
        setFormValues((prev) => ({ ...prev, [key]: val }));
    };

    const handleCheckboxToggle = (key: string, optionValue: string) => {
        setFormValues((prev) => {
            const current = (prev[key] as string[]) || [];
            if (current.includes(optionValue)) {
                return { ...prev, [key]: current.filter((v) => v !== optionValue) };
            }
            return { ...prev, [key]: [...current, optionValue] };
        });
    };

    const handleSwitchChange = (key: string, checked: boolean) => {
        setFormValues((prev) => ({ ...prev, [key]: checked }));
    };

    return (
        <div className="flex flex-col h-full min-h-0 overflow-y-auto p-4 sm:p-6 lg:p-8 bg-muted/20">
            <div className="mx-auto w-full max-w-xl space-y-6">
                {/* 15. Event Details on the top */}
                <div className="rounded-xl border bg-card p-5 shadow-xs space-y-3">
                    <div className="flex items-center justify-between gap-2">
                        <Badge variant="outline" className="text-[11px] gap-1 px-2 py-0.5 font-medium border-primary/30 text-primary bg-primary/5">
                            <ShieldCheck className="size-3 text-primary" />
                            <span>Official Event Form</span>
                        </Badge>
                        <span className="text-[11px] text-muted-foreground font-mono">
                            Live Preview
                        </span>
                    </div>

                    <div>
                        <h2 className="text-lg sm:text-xl font-bold tracking-tight text-foreground">
                            {form?.name || "Untitled Form"}
                        </h2>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-2 border-t text-[11px] text-muted-foreground">
                        <div className="flex items-center gap-2">
                            <Calendar className="size-3.5 text-primary shrink-0" />
                            <span>Scheduled: Multi-Date Access</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <Clock className="size-3.5 text-primary shrink-0" />
                            <span>Standard Event Hours</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <MapPin className="size-3.5 text-primary shrink-0" />
                            <span>Main Campus / Venue</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <UserCheck className="size-3.5 text-primary shrink-0" />
                            <span>Identity Card Verified</span>
                        </div>
                    </div>
                </div>

                {/* Actual Form Fields Card */}
                <div className="rounded-xl border bg-card p-5 sm:p-6 shadow-xs space-y-5">
                    <div className="border-b pb-3">
                        <h3 className="text-sm font-semibold text-foreground">
                            Attendee Registration Information
                        </h3>
                        <p className="text-[11px] text-muted-foreground mt-0.5">
                            Please fill out all mandatory fields marked with an asterisk (*).
                        </p>
                    </div>

                    <div className="space-y-4">
                        {
                            fields.map((field) => {
                                const val = (formValues[field.key] as string) || "";

                                return (
                                    <div key={field.id} className="space-y-1.5">
                                        <div className="flex items-center justify-between">
                                            <Label htmlFor={`preview-${field.id}`} className="text-xs font-semibold text-foreground">
                                                {field.label}
                                                {
                                                    field.required && <span className="text-destructive ml-1">*</span>
                                                }
                                            </Label>
                                            {
                                                field.is_system && <span className="text-[10px] text-muted-foreground font-mono">
                                                    Required
                                                </span>
                                            }
                                        </div>

                                        {/* Field Renderers for 12 types */}
                                        {
                                            (field.type === "text" || field.type === "email" || field.type === "url") && (
                                                <Input
                                                    id={`preview-${field.id}`}
                                                    type={field.type}
                                                    value={val}
                                                    onChange={(e) => handleTextChange(field.key, e.target.value)}
                                                    placeholder={field.placeholder || undefined}
                                                    className="text-xs h-9 bg-background"
                                                />
                                            )
                                        }

                                        {
                                            field.type === "phone" && (
                                                <PhoneInput
                                                    placeholder={field.placeholder || "10-digit mobile number"}
                                                    value={val}
                                                    onChange={(phoneVal) => handleTextChange(field.key, phoneVal)}
                                                />
                                            )
                                        }

                                        {
                                            field.type === "textarea" && (
                                                <Textarea
                                                    id={`preview-${field.id}`}
                                                    value={val}
                                                    onChange={(e) => handleTextChange(field.key, e.target.value)}
                                                    placeholder={field.placeholder || undefined}
                                                    className="text-xs min-h-[80px] bg-background"
                                                />
                                            )
                                        }

                                        {
                                            field.type === "number" && (
                                                <Input
                                                    id={`preview-${field.id}`}
                                                    type="number"
                                                    value={val}
                                                    min={field.validation?.min ?? undefined}
                                                    max={field.validation?.max ?? undefined}
                                                    onChange={(e) => handleTextChange(field.key, e.target.value)}
                                                    placeholder={field.placeholder || undefined}
                                                    className="text-xs h-9 bg-background"
                                                />
                                            )
                                        }

                                        {
                                            (field.type === "date" || field.type === "time" || field.type === "month" || field.type === "week") && (
                                                <Input
                                                    id={`preview-${field.id}`}
                                                    type={field.type}
                                                    value={val}
                                                    onChange={(e) => handleTextChange(field.key, e.target.value)}
                                                    className="text-xs h-9 bg-background"
                                                />
                                            )
                                        }

                                        {
                                            field.type === "switch" && (
                                                <div className="flex items-center gap-2 pt-1">
                                                    <Switch
                                                        id={`preview-${field.id}`}
                                                        checked={Boolean(formValues[field.key])}
                                                        onCheckedChange={(checked) => handleSwitchChange(field.key, checked)}
                                                    />
                                                    <span className="text-xs text-muted-foreground">
                                                        {field.placeholder || "Yes / Enable"}
                                                    </span>
                                                </div>
                                            )
                                        }

                                        {
                                            field.type === "checkbox" && (
                                                <div className="space-y-2 pt-1">
                                                    {
                                                        field.options?.map((opt) => {
                                                            const isChecked = ((formValues[field.key] as string[]) || []).includes(opt.value);
                                                            return (
                                                                <div key={opt.id} className="flex items-center gap-2">
                                                                    <Checkbox
                                                                        id={`preview-chk-${opt.id}`}
                                                                        checked={isChecked}
                                                                        onCheckedChange={() => handleCheckboxToggle(field.key, opt.value)}
                                                                    />
                                                                    <Label
                                                                        htmlFor={`preview-chk-${opt.id}`}
                                                                        className="text-xs font-normal text-foreground cursor-pointer"
                                                                    >
                                                                        {opt.label}
                                                                    </Label>
                                                                </div>
                                                            );
                                                        })
                                                    }
                                                </div>
                                            )
                                        }

                                        {
                                            field.type === "radio" && (
                                                <div className="space-y-2 pt-1">
                                                    {
                                                        field.options?.map((opt) => {
                                                            return (
                                                                <div key={opt.id} className="flex items-center gap-2">
                                                                    <input
                                                                        type="radio"
                                                                        id={`preview-rad-${opt.id}`}
                                                                        name={`radio-${field.id}`}
                                                                        value={opt.value}
                                                                        checked={val === opt.value}
                                                                        onChange={() => handleTextChange(field.key, opt.value)}
                                                                        className="size-3.5 accent-primary cursor-pointer"
                                                                    />
                                                                    <Label
                                                                        htmlFor={`preview-rad-${opt.id}`}
                                                                        className="text-xs font-normal text-foreground cursor-pointer"
                                                                    >
                                                                        {opt.label}
                                                                    </Label>
                                                                </div>
                                                            );
                                                        })
                                                    }
                                                </div>
                                            )
                                        }

                                        {
                                            field.type === "file" && (
                                                <div className="flex flex-col items-center justify-center gap-2 rounded-lg border border-dashed border-input p-5 text-center bg-muted/10 hover:bg-muted/20 transition-colors cursor-pointer">
                                                    <div className="flex size-9 items-center justify-center rounded-full bg-muted">
                                                        <UploadCloud className="size-4 text-muted-foreground" />
                                                    </div>
                                                    <div className="flex flex-col gap-0.5">
                                                        <span className="text-xs font-medium text-foreground">
                                                            {
                                                                field.validation?.multiple ? "Click or drag multiple files to upload" : "Click or drag file to upload"
                                                            }
                                                        </span>
                                                        <span className="text-[11px] text-muted-foreground">
                                                            {
                                                                field.validation?.accept ? `Accepted: ${field.validation.accept}` : "All common files accepted"
                                                            }
                                                            {
                                                                field.validation?.max_file_size_mb && ` • Max: ${field.validation.max_file_size_mb} MB`
                                                            }
                                                            {
                                                                field.validation?.multiple && " • Multiple files allowed"
                                                            }
                                                        </span>
                                                    </div>
                                                </div>
                                            )
                                        }
                                    </div>
                                );
                            })
                        }
                    </div>

                    {/* 15. Footer: Cancel and Submit button */}
                    <div className="flex items-center justify-end gap-2.5 pt-4 border-t">
                        <Button
                            type="button"
                            variant="outline"
                            className="text-xs"
                            onClick={() => setFormValues({})}
                        >
                            Cancel
                        </Button>
                        <Button
                            type="button"
                            className="text-xs font-semibold px-5"
                            onClick={() => alert("Preview mode: Form response validated.")}
                        >
                            Submit Application
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
}
