"use client";

import * as React from "react";
import { AlertTriangle, Loader2, Send } from "lucide-react";

import {
    AlertDialog,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { usePublishFormMutation } from "@/query-hooks/forms.api";

interface PublishFormDialogProps {
    formId: string;
    formName: string;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function PublishFormDialog({
    formId,
    formName,
    open,
    onOpenChange,
}: PublishFormDialogProps) {
    const publishMutation = usePublishFormMutation();

    const handleConfirm = async () => {
        try {
            await publishMutation.mutateAsync(formId);
            onOpenChange(false);
        } catch {
            // error handled by mutation toast
        }
    };

    return (
        <AlertDialog open={open} onOpenChange={onOpenChange}>
            <AlertDialogContent className="max-w-md">
                <AlertDialogHeader>
                    <div className="flex items-center gap-2.5 text-amber-500">
                        <div className="flex size-9 items-center justify-center rounded-full bg-amber-500/10">
                            <AlertTriangle className="size-5" />
                        </div>
                        <AlertDialogTitle className="text-base font-semibold">
                            Publish &ldquo;{formName}&rdquo;?
                        </AlertDialogTitle>
                    </div>
                    <AlertDialogDescription className="text-xs text-muted-foreground pt-2 leading-relaxed">
                        Once published, this form structure will be permanently locked and cannot be edited, reordered, or deleted. This ensures applicant data integrity when assigned to events.
                    </AlertDialogDescription>
                </AlertDialogHeader>

                <AlertDialogFooter className="pt-3">
                    <AlertDialogCancel disabled={publishMutation.isPending} className="text-xs">
                        Cancel
                    </AlertDialogCancel>
                    <Button
                        type="button"
                        onClick={handleConfirm}
                        disabled={publishMutation.isPending}
                        className="gap-2 text-xs font-semibold bg-emerald-600 hover:bg-emerald-700 text-white"
                    >
                        {publishMutation.isPending ? (
                            <Loader2 className="size-3.5 animate-spin" />
                        ) : (
                            <Send className="size-3.5" />
                        )}
                        <span>{publishMutation.isPending ? "Publishing..." : "Confirm & Publish"}</span>
                    </Button>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
}
