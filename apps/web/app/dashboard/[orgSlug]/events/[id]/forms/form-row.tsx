"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
    Card,
    CardAction,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useDeleteFormMutation, useUpdateFormMutation } from "@/query-hooks/forms.api";
import type { EventForm } from "@/schema/forms.types";

const APP_URL = process.env.NEXT_PUBLIC_APP_URL ?? "";

export function FormRow({
    eventId,
    form,
    subEventName,
}: {
    eventId: string;
    form: EventForm;
    subEventName?: string;
}) {
    const updateMutation = useUpdateFormMutation(eventId, form.id);
    const deleteMutation = useDeleteFormMutation(eventId);
    const [copied, setCopied] = useState(false);

    const link = `${APP_URL}/forms/${form.token}`;

    const copyLink = async () => {
        await navigator.clipboard.writeText(link);
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
    };

    const toggleActive = async () => {
        await updateMutation.execute({
            capacity: form.capacity ?? undefined,
            is_active: !form.is_active,
        });
    };

    const remove = async () => {
        await deleteMutation.execute(form.id);
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle>{subEventName ?? "Whole event"}</CardTitle>
                <CardDescription>
                    {form.submissions_count}
                    {form.capacity ? ` / ${form.capacity}` : ""} submissions ·{" "}
                    {form.is_active ? "Active" : "Inactive"}
                </CardDescription>
                <CardAction className="flex items-center gap-2">
                    <Button variant="outline" size="sm" onClick={copyLink}>
                        {copied ? "Copied" : "Copy link"}
                    </Button>
                    <Button
                        variant="outline"
                        size="sm"
                        disabled={updateMutation.isPending}
                        onClick={toggleActive}
                    >
                        {form.is_active ? "Deactivate" : "Activate"}
                    </Button>
                    <Button
                        variant="destructive"
                        size="sm"
                        disabled={deleteMutation.isPending}
                        onClick={remove}
                    >
                        Delete
                    </Button>
                </CardAction>
            </CardHeader>
            <CardContent className="flex flex-col gap-1">
                <Input
                    readOnly
                    value={link}
                    onFocus={(e) => e.currentTarget.select()}
                    className="text-xs"
                />
                {updateMutation.error && (
                    <p className="text-xs text-destructive">{updateMutation.error.message}</p>
                )}
                {deleteMutation.error && (
                    <p className="text-xs text-destructive">{deleteMutation.error.message}</p>
                )}
            </CardContent>
        </Card>
    );
}