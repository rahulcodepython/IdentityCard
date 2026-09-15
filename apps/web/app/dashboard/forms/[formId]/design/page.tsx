"use client";

import * as React from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { Check, Eye, FileText, Loader2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FormDesignerCanvas } from "@/components/forms/designer/form-designer-canvas";
import { FormPreviewPanel } from "@/components/forms/preview/form-preview-panel";
import { useBreadcrumbs } from "@/hooks/use-breadcrumbs";
import { useFormQuery, useUpdateFormFieldsMutation } from "@/query-hooks/forms.api";
import type { FormField } from "@/schema/forms.types";

export default function DashboardFormDesignPage() {
    const params = useParams();
    const formId = params?.formId as string;

    const { data: form, isLoading, isError, error } = useFormQuery(formId);
    const updateFieldsMutation = useUpdateFormFieldsMutation();

    const [fields, setFields] = React.useState<FormField[]>([]);
    const [initialLoaded, setInitialLoaded] = React.useState(false);

    // Top navbar breadcrumb updates
    useBreadcrumbs([
        {
            title: "Dashboard",
            url: "/dashboard",
        },
        {
            title: "Forms",
            url: "/dashboard/forms",
        },
        {
            title: form?.name || "Form",
        },
        {
            title: "Design",
        },
    ]);

    // Synchronize loaded form fields to local builder state
    React.useEffect(() => {
        if (form && !initialLoaded) {
            setFields(form.fields || []);
            setInitialLoaded(true);
        }
    }, [form, initialLoaded]);

    const hasUnsavedChanges = React.useMemo(() => {
        if (!form) return false;
        return JSON.stringify(form.fields) !== JSON.stringify(fields);
    }, [form, fields]);

    const handleSaveFields = async () => {
        if (!formId) return;
        await updateFieldsMutation.mutateAsync({
            id: formId,
            fields,
        });
    };

    if (isLoading) {
        return (
            <div className="flex h-[calc(100vh-6rem)] w-full items-center justify-center">
                <div className="flex flex-col items-center gap-2">
                    <Loader2 className="size-6 animate-spin text-primary" />
                    <span className="text-xs text-muted-foreground">Loading form designer...</span>
                </div>
            </div>
        );
    }

    if (isError || !form) {
        return (
            <div className="flex h-[calc(100vh-6rem)] w-full flex-col items-center justify-center gap-3 p-4 text-center">
                <span className="text-sm font-semibold text-destructive">
                    {error?.message || "Form not found"}
                </span>
                <Button
                    nativeButton={false}
                    render={<Link href="/dashboard/forms" />}
                    variant="outline"
                    className="text-xs"
                >
                    Back to Forms
                </Button>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-[calc(100vh-5.5rem)] min-h-[calc(100vh-5.5rem)] max-h-[calc(100vh-5.5rem)] w-full gap-3 overflow-hidden">
            {/* Action Bar */}
            <div className="flex items-center justify-between px-1 shrink-0">
                <div className="flex items-center gap-2.5 min-w-0">
                    <div className="flex size-7 items-center justify-center rounded-md bg-primary/10 text-primary">
                        <FileText className="size-4" />
                    </div>
                    <div className="flex items-center gap-2 min-w-0">
                        <h1 className="text-sm sm:text-base font-bold text-foreground truncate">
                            {form.name}
                        </h1>
                        <Badge variant="outline" className="text-[10px] hidden sm:inline-flex">
                            Designer
                        </Badge>
                    </div>
                </div>

                <div className="flex items-center gap-3">
                    <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                        {
                            hasUnsavedChanges ? (
                                <span className="flex items-center gap-1 text-amber-500 dark:text-amber-400 font-medium text-[11px]">
                                    <span className="size-1.5 rounded-full bg-amber-500 animate-pulse" />
                                    <span>Unsaved changes</span>
                                </span>
                            ) : (
                                <span className="flex items-center gap-1 text-muted-foreground text-[11px]">
                                    <Check className="size-3 text-emerald-500" />
                                    <span>Saved to cloud</span>
                                </span>
                            )
                        }
                    </div>

                    <div className="hidden sm:flex items-center gap-1.5 rounded-md border bg-muted/30 px-2.5 py-1 text-[11px] text-muted-foreground">
                        <Eye className="size-3.5 text-primary" />
                        <span>Live Sync Preview</span>
                    </div>
                </div>
            </div>

            {/* 7 & 8. Split the page into two same width partitions with max/min height screen */}
            <main className="flex-1 grid grid-cols-1 lg:grid-cols-2 min-h-0 overflow-hidden rounded-xl border divide-y lg:divide-y-0 lg:divide-x bg-card shadow-xs">
                {/* First part: Form edit section */}
                <section className="h-full min-h-0 overflow-hidden flex flex-col">
                    <FormDesignerCanvas
                        form={form}
                        fields={fields}
                        onFieldsChange={setFields}
                        onSave={handleSaveFields}
                        isSaving={updateFieldsMutation.isPending}
                        hasUnsavedChanges={hasUnsavedChanges}
                    />
                </section>

                {/* Second part: Form preview section */}
                <section className="h-full min-h-0 overflow-hidden flex flex-col">
                    <FormPreviewPanel form={form} fields={fields} />
                </section>
            </main>
        </div>
    );
}
