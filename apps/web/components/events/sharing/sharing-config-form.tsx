"use client";

import * as React from "react";
import Link from "next/link";
import { zodResolver } from "@hookform/resolvers/zod";
import { FileText, Loader2, Save } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import {
    Form,
    FormControl,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { useUpdateEventSharingMutation } from "@/query-hooks/eventsharing.api";
import { useFormsInfiniteQuery } from "@/query-hooks/forms.api";
import {
    SharingConfigFormSchema,
    type EventSharingResponse,
    type SharingConfigFormValues,
} from "@/schema/eventsharing.types";

interface SharingConfigFormProps {
    eventId: string;
    sharing?: EventSharingResponse;
}

function getFormattedExpiresAt(dateStr?: string | null): string {
    if (!dateStr) return "";
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return "";
    const tzOffset = d.getTimezoneOffset() * 60000;
    return new Date(d.getTime() - tzOffset).toISOString().slice(0, 16);
}

export function SharingConfigForm({ eventId, sharing }: SharingConfigFormProps) {
    const updateSharingMutation = useUpdateEventSharingMutation();

    const hasAssignedForm = Boolean(sharing?.event_form);
    const [isChangingForm, setIsChangingForm] = React.useState(false);

    const form = useForm<SharingConfigFormValues>({
        resolver: zodResolver(SharingConfigFormSchema),
        defaultValues: {
            form_id: sharing?.event_form?.form_id || "",
            max_applicants: sharing?.event_form?.max_applicants ?? -1,
            expires_at: getFormattedExpiresAt(sharing?.event_form?.expires_at),
            status: sharing?.event_form?.status || "waiting",
        },
    });

    // Synchronize if sharing data changes from server (e.g. on initial fetch or refetch)
    const syncedFormIdRef = React.useRef(sharing?.event_form?.id);
    React.useEffect(() => {
        if (sharing?.event_form && sharing.event_form.id !== syncedFormIdRef.current) {
            syncedFormIdRef.current = sharing.event_form.id;
            form.reset({
                form_id: sharing.event_form.form_id,
                max_applicants: sharing.event_form.max_applicants ?? -1,
                expires_at: getFormattedExpiresAt(sharing.event_form.expires_at),
                status: sharing.event_form.status || "waiting",
            });
        }
    }, [sharing?.event_form, form]);

    // Only fetch published forms if user has no form or is changing form
    const shouldFetchForms = !hasAssignedForm || isChangingForm;
    const { data: formsData, isLoading: isFormsLoading } = useFormsInfiniteQuery({
        published: true,
        enabled: shouldFetchForms,
    });

    const publishedForms = React.useMemo(() => {
        return formsData?.pages.flatMap((p) => p.data) || [];
    }, [formsData]);

    const canChangeForm = sharing?.can_change_form ?? true;

    const onSubmit = async (values: SharingConfigFormValues) => {
        const expiryDate = new Date(values.expires_at);

        try {
            await updateSharingMutation.mutateAsync({
                eventId,
                form_id: values.form_id,
                max_applicants: values.max_applicants,
                expires_at: expiryDate.toISOString(),
                status: values.status,
            });
            setIsChangingForm(false);
        } catch {
            // Error toast handled by mutation onError
        }
    };

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)}>
                <Card className="shadow-xs">
                    <CardHeader className="pb-4">
                        <CardTitle className="text-sm font-semibold flex items-center gap-2">
                            <FileText className="size-4 text-primary" />
                            <span>Registration Form Assignment</span>
                        </CardTitle>
                        <CardDescription className="text-xs">
                            Select which published form collects applicant responses for this event.
                        </CardDescription>
                    </CardHeader>

                    <CardContent className="space-y-4">
                        {/* Active Assigned Form Banner */}
                        {hasAssignedForm && !isChangingForm ? (
                            <div className="rounded-lg border bg-card p-4 space-y-3">
                                <div className="flex items-start justify-between gap-3">
                                    <div>
                                        <div className="flex items-center gap-2">
                                            <span className="text-xs font-semibold text-foreground">
                                                {sharing?.assigned_form?.name || "Assigned Form"}
                                            </span>
                                        </div>
                                        <p className="text-[11px] text-muted-foreground mt-0.5">
                                            {sharing?.assigned_form?.fields_count || 0} fields configured
                                        </p>
                                    </div>

                                    {canChangeForm ? (
                                        <Button
                                            type="button"
                                            variant="outline"
                                            size="sm"
                                            onClick={() => setIsChangingForm(true)}
                                            className="text-xs"
                                        >
                                            Change Form
                                        </Button>
                                    ) : (
                                        <span className="text-[10px] text-muted-foreground italic">
                                            Cannot change (responses recorded)
                                        </span>
                                    )}
                                </div>
                            </div>
                        ) : (
                            <div className="space-y-2">
                                <FormField
                                    control={form.control}
                                    name="form_id"
                                    render={({ field }) => (
                                        <FormItem className="space-y-1.5">
                                            <FormLabel className="text-xs font-medium">
                                                Choose a Published Form
                                            </FormLabel>
                                            {isFormsLoading ? (
                                                <div className="flex items-center gap-2 text-xs text-muted-foreground p-2">
                                                    <Loader2 className="size-3.5 animate-spin" />
                                                    <span>Loading published forms...</span>
                                                </div>
                                            ) : publishedForms.length === 0 ? (
                                                <div className="p-3 rounded-lg border border-dashed text-xs text-muted-foreground flex flex-col gap-2">
                                                    <span>
                                                        No published forms available. You must design and publish a form first.
                                                    </span>
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        nativeButton={false}
                                                        render={<Link href="/dashboard/forms" />}
                                                        className="w-fit text-xs"
                                                    >
                                                        Manage Forms
                                                    </Button>
                                                </div>
                                            ) : (
                                                <FormControl>
                                                    <select
                                                        {...field}
                                                        className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-xs shadow-xs focus:outline-none focus:ring-1 focus:ring-ring"
                                                    >
                                                        <option value="">Select a form...</option>
                                                        {publishedForms.map((f) => (
                                                            <option key={f.id} value={f.id}>
                                                                {f.name} ({f.fields?.length || 0} fields)
                                                            </option>
                                                        ))}
                                                    </select>
                                                </FormControl>
                                            )}
                                            <FormMessage className="text-[11px]" />
                                        </FormItem>
                                    )}
                                />

                                {isChangingForm && (
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => {
                                            form.setValue(
                                                "form_id",
                                                sharing?.event_form?.form_id || ""
                                            );
                                            setIsChangingForm(false);
                                        }}
                                        className="text-xs text-muted-foreground"
                                    >
                                        Cancel change
                                    </Button>
                                )}
                            </div>
                        )}

                        {/* Registration Rules */}
                        <div className="pt-2 border-t space-y-4">
                            <span className="text-xs font-semibold text-foreground block">
                                Registration Capacity & Deadline
                            </span>

                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                <FormField
                                    control={form.control}
                                    name="max_applicants"
                                    render={({ field }) => (
                                        <FormItem className="space-y-1.5">
                                            <FormLabel className="text-xs font-medium">
                                                Max Applicants
                                            </FormLabel>
                                            <FormControl>
                                                <Input
                                                    type="number"
                                                    placeholder="-1 for unlimited"
                                                    className="text-xs"
                                                    {...field}
                                                    value={field.value ?? ""}
                                                    onChange={(e) =>
                                                        field.onChange(
                                                            e.target.value === ""
                                                                ? ""
                                                                : Number(e.target.value)
                                                        )
                                                    }
                                                />
                                            </FormControl>
                                            <FormDescription className="text-[10px] text-muted-foreground">
                                                Enter -1 for unlimited capacity.
                                            </FormDescription>
                                            <FormMessage className="text-[11px]" />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="expires_at"
                                    render={({ field }) => (
                                        <FormItem className="space-y-1.5">
                                            <FormLabel className="text-xs font-medium">
                                                Registration Deadline
                                            </FormLabel>
                                            <FormControl>
                                                <Input
                                                    type="datetime-local"
                                                    className="text-xs"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormDescription className="text-[10px] text-muted-foreground">
                                                Applications close automatically at this date/time.
                                            </FormDescription>
                                            <FormMessage className="text-[11px]" />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            {/* Status Toggle */}
                            <FormField
                                control={form.control}
                                name="status"
                                render={({ field }) => (
                                    <FormItem className="flex items-center justify-between p-3 rounded-lg border bg-muted/20">
                                        <div className="space-y-0.5">
                                            <FormLabel className="text-xs font-medium">
                                                Accept Applications (Live)
                                            </FormLabel>
                                            <FormDescription className="text-[10px] text-muted-foreground block">
                                                {field.value === "live"
                                                    ? "Public registration is currently active and open."
                                                    : "Form is paused. Visitors will see a waiting message."}
                                            </FormDescription>
                                        </div>
                                        <FormControl>
                                            <Switch
                                                checked={field.value === "live"}
                                                onCheckedChange={(checked) =>
                                                    field.onChange(checked ? "live" : "waiting")
                                                }
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />
                        </div>
                    </CardContent>

                    <CardFooter className="flex items-center justify-end pt-4 border-t">
                        <Button
                            type="submit"
                            disabled={updateSharingMutation.isPending}
                            className="gap-2 text-xs font-semibold"
                        >
                            {updateSharingMutation.isPending ? (
                                <Loader2 className="size-3.5 animate-spin" />
                            ) : (
                                <Save className="size-3.5" />
                            )}
                            <span>
                                {updateSharingMutation.isPending
                                    ? "Saving..."
                                    : "Save Configuration"}
                            </span>
                        </Button>
                    </CardFooter>
                </Card>
            </form>
        </Form>
    );
}
