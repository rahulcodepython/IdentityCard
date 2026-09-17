"use client";

import * as React from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2, UserPlus } from "lucide-react";
import { useForm } from "react-hook-form";

import { ApplyFieldInput } from "@/components/apply/apply-field-input";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PhoneInput } from "@/components/ui/phone-input";
import { useCreateApplicantMutation } from "@/query-hooks/applicants.api";
import type { FormFieldSummary } from "@/schema/applicants.types";
import type { FormField as FormFieldType } from "@/schema/forms.types";
import {
    buildPublicApplyFormSchema,
    type FormFieldValue,
    type PublicApplyFormValues,
} from "@/schema/publicapply.types";

interface CreateApplicantDialogProps {
    eventId: string;
    formFields?: FormFieldSummary[];
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function CreateApplicantDialog({
    eventId,
    formFields = [],
    open,
    onOpenChange,
}: CreateApplicantDialogProps) {
    const createMutation = useCreateApplicantMutation(eventId);

    const customFields = React.useMemo(() => {
        return formFields.map(
            (f) =>
                ({
                    id: f.id || f.key,
                    key: f.key,
                    label: f.label,
                    type: f.type as FormFieldType["type"],
                    required: f.required,
                    placeholder: f.placeholder,
                    options: f.options,
                }) as FormFieldType
        );
    }, [formFields]);

    const schema = React.useMemo(
        () => buildPublicApplyFormSchema(customFields),
        [customFields]
    );

    const form = useForm<PublicApplyFormValues>({
        resolver: zodResolver(schema),
        defaultValues: {
            name: "",
            email: "",
            phone: "",
            data: {},
        },
    });

    React.useEffect(() => {
        if (open) {
            form.reset({
                name: "",
                email: "",
                phone: "",
                data: {},
            });
        }
    }, [open, form]);

    const onSubmit = async (values: PublicApplyFormValues) => {
        try {
            const cleanData: Record<string, FormFieldValue> = {
                phone: values.phone.trim(),
            };
            for (const [k, v] of Object.entries(values.data || {})) {
                if (v !== undefined) {
                    cleanData[k] = v;
                }
            }

            await createMutation.mutateAsync({
                name: values.name.trim(),
                email: values.email.trim().toLowerCase(),
                data: cleanData,
            });

            onOpenChange(false);
        } catch {
            // Handled by mutation toast
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
                <DialogHeader>
                    <div className="flex items-center gap-2">
                        <div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                            <UserPlus className="size-4" />
                        </div>
                        <div>
                            <DialogTitle className="text-base font-semibold">
                                Add Applicant Manually
                            </DialogTitle>
                            <DialogDescription className="text-xs">
                                Register a new attendee for this event using the configured registration form.
                            </DialogDescription>
                        </div>
                    </div>
                </DialogHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 pt-2">
                        <FormField
                            control={form.control}
                            name="name"
                            render={({ field }) => (
                                <FormItem className="space-y-1.5">
                                    <FormLabel className="text-xs font-medium">
                                        Full Name{" "}
                                        <span className="text-destructive font-bold">*</span>
                                    </FormLabel>
                                    <FormControl>
                                        <Input
                                            placeholder="Enter attendee full name"
                                            className="text-xs h-9"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage className="text-[11px]" />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="email"
                            render={({ field }) => (
                                <FormItem className="space-y-1.5">
                                    <FormLabel className="text-xs font-medium">
                                        Email Address{" "}
                                        <span className="text-destructive font-bold">*</span>
                                    </FormLabel>
                                    <FormControl>
                                        <Input
                                            type="email"
                                            placeholder="attendee.email@example.com"
                                            className="text-xs h-9"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage className="text-[11px]" />
                                </FormItem>
                            )}
                        />

                        {/* Mandatory Mobile Number */}
                        <FormField
                            control={form.control}
                            name="phone"
                            render={({ field }) => (
                                <FormItem className="space-y-1.5">
                                    <FormLabel className="text-xs font-medium">
                                        Mobile Number{" "}
                                        <span className="text-destructive font-bold">*</span>
                                    </FormLabel>
                                    <FormControl>
                                        <PhoneInput
                                            value={field.value}
                                            onChange={field.onChange}
                                            placeholder="10-digit mobile number"
                                        />
                                    </FormControl>
                                    <FormMessage className="text-[11px]" />
                                </FormItem>
                            )}
                        />

                        {
                            customFields.map((field) => {
                                if (
                                    field.is_system &&
                                    (field.key === "name" ||
                                        field.key === "email" ||
                                        field.key === "phone")
                                ) {
                                    return null;
                                }

                                return <ApplyFieldInput
                                    key={field.id || field.key}
                                    field={field}
                                    control={form.control}
                                />;
                            })
                        }

                        <div className="flex items-center justify-end gap-2 pt-3 border-t">
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={() => onOpenChange(false)}
                                className="text-xs"
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                size="sm"
                                disabled={createMutation.isPending}
                                className="text-xs gap-1.5"
                            >
                                {
                                    createMutation.isPending ? <Loader2 className="size-3.5 animate-spin" /> : <UserPlus className="size-3.5" />
                                }
                                <span>
                                    {
                                        createMutation.isPending ? "Adding Attendee..." : "Add Attendee"
                                    }
                                </span>
                            </Button>
                        </div>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
}
