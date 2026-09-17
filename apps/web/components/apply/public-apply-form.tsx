"use client";

import * as React from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2, Send } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { ApplyFieldInput } from "@/components/apply/apply-field-input";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
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
import { useSubmitApplicationMutation } from "@/query-hooks/publicapply.api";
import type { FormField as FormFieldType } from "@/schema/forms.types";
import {
    buildPublicApplyFormSchema,
    type FormFieldValue,
    type PublicApplyConfig,
    type PublicApplyFormValues,
    type SubmitApplicationResponse,
} from "@/schema/publicapply.types";

interface PublicApplyFormProps {
    eventFormId: string;
    config: PublicApplyConfig;
    onSuccess: (res: SubmitApplicationResponse, submittedEmail: string) => void;
}

export function PublicApplyForm({
    eventFormId,
    config,
    onSuccess,
}: PublicApplyFormProps) {
    const submitMutation = useSubmitApplicationMutation();
    const customFields = (config.form?.fields as FormFieldType[]) || [];

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

    const onSubmit = async (values: PublicApplyFormValues) => {
        try {
            const parsed = schema.parse(values);
            const cleanData: Record<string, FormFieldValue> = {
                phone: parsed.phone.trim(),
            };
            for (const [k, v] of Object.entries(parsed.data || {})) {
                if (v !== undefined) {
                    cleanData[k] = v;
                }
            }
            const res = await submitMutation.mutateAsync({
                eventFormId,
                name: parsed.name.trim(),
                email: parsed.email.trim().toLowerCase(),
                data: cleanData,
            });
            onSuccess(res, parsed.email.trim().toLowerCase());
            toast.success("Application submitted successfully!");
        } catch {
            // Handled by mutation onError toast
        }
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-base font-semibold">
                    Attendee Information
                </CardTitle>
                <CardDescription className="text-xs">
                    Fields marked with an asterisk (
                    <span className="text-destructive font-bold">*</span>) are
                    mandatory.
                </CardDescription>
            </CardHeader>

            <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)}>
                    <CardContent className="space-y-4">
                        {/* Mandatory Name */}
                        <FormField
                            control={form.control}
                            name="name"
                            render={({ field }) => (
                                <FormItem className="space-y-1.5">
                                    <FormLabel className="text-xs font-medium">
                                        Full Name{" "}
                                        <span className="text-destructive font-bold">
                                            *
                                        </span>
                                    </FormLabel>
                                    <FormControl>
                                        <Input
                                            placeholder="Enter your full name"
                                            className="text-xs h-9"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage className="text-[11px]" />
                                </FormItem>
                            )}
                        />

                        {/* Mandatory Email */}
                        <FormField
                            control={form.control}
                            name="email"
                            render={({ field }) => (
                                <FormItem className="space-y-1.5">
                                    <FormLabel className="text-xs font-medium">
                                        Email Address{" "}
                                        <span className="text-destructive font-bold">
                                            *
                                        </span>
                                    </FormLabel>
                                    <FormControl>
                                        <Input
                                            type="email"
                                            placeholder="your.email@example.com"
                                            className="text-xs h-9"
                                            {...field}
                                        />
                                    </FormControl>
                                    <p className="text-[10px] text-muted-foreground">
                                        Only one registration per email is allowed for this event.
                                    </p>
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
                                        <span className="text-destructive font-bold">
                                            *
                                        </span>
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

                        {/* Dynamic Custom Fields */}
                        {customFields.map((field) => {
                            if (
                                field.is_system &&
                                (field.key === "name" ||
                                    field.key === "email" ||
                                    field.key === "phone")
                            ) {
                                return null;
                            }

                            return (
                                <ApplyFieldInput
                                    key={field.id}
                                    field={field}
                                    control={form.control}
                                />
                            );
                        })}

                        <div className="pt-2">
                            <Button
                                type="submit"
                                disabled={submitMutation.isPending}
                                className="w-full gap-2 text-xs font-semibold h-10"
                            >
                                {submitMutation.isPending ? (
                                    <Loader2 className="size-4 animate-spin" />
                                ) : (
                                    <Send className="size-4" />
                                )}
                                <span>
                                    {submitMutation.isPending
                                        ? "Submitting Application..."
                                        : "Submit Registration"}
                                </span>
                            </Button>
                        </div>
                    </CardContent>
                </form>
            </Form>
        </Card>
    );
}
