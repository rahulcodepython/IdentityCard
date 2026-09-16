"use client";

import * as React from "react";
import { Eye, Loader2, MoreHorizontal, QrCode, Trash2 } from "lucide-react";

import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useDeleteApplicantMutation } from "@/query-hooks/applicants.api";
import type { ApplicantItem, FormFieldSummary } from "@/schema/applicants.types";
import { ApplicantDataDialog } from "./applicant-data-dialog";
import { ApplicantQrDialog } from "./applicant-qr-dialog";

interface ApplicantRowActionsProps {
    applicant: ApplicantItem;
    formFields?: FormFieldSummary[];
    eventId?: string;
}

export function ApplicantRowActions({
    applicant,
    formFields,
    eventId = "",
}: ApplicantRowActionsProps) {
    const [detailsOpen, setDetailsOpen] = React.useState(false);
    const [qrOpen, setQrOpen] = React.useState(false);
    const [deleteConfirmOpen, setDeleteConfirmOpen] = React.useState(false);

    const deleteMutation = useDeleteApplicantMutation(eventId);

    const handleDelete = async () => {
        try {
            await deleteMutation.mutateAsync(applicant.user_id);
            setDeleteConfirmOpen(false);
        } catch {
            // Handled by mutation toast
        }
    };

    return (
        <div className="flex items-center justify-end">
            <DropdownMenu>
                <DropdownMenuTrigger
                    render={
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-8 text-muted-foreground hover:text-foreground"
                        >
                            <MoreHorizontal className="size-4" />
                            <span className="sr-only">Open menu</span>
                        </Button>
                    }
                />
                <DropdownMenuContent align="end" className="w-44 text-xs">
                    <DropdownMenuItem
                        onClick={() => setDetailsOpen(true)}
                        className="gap-2 cursor-pointer"
                    >
                        <Eye className="size-3.5 text-muted-foreground" />
                        <span>View Form Data</span>
                    </DropdownMenuItem>

                    <DropdownMenuItem
                        onClick={() => setQrOpen(true)}
                        className="gap-2 cursor-pointer"
                    >
                        <QrCode className="size-3.5 text-primary" />
                        <span>View QR Code</span>
                    </DropdownMenuItem>

                    <DropdownMenuSeparator />

                    <DropdownMenuItem
                        onClick={() => setDeleteConfirmOpen(true)}
                        className="gap-2 text-destructive focus:text-destructive cursor-pointer"
                    >
                        <Trash2 className="size-3.5" />
                        <span>Delete Applicant</span>
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            <ApplicantDataDialog
                applicant={applicant}
                formFields={formFields}
                open={detailsOpen}
                onOpenChange={setDetailsOpen}
            />

            <ApplicantQrDialog
                applicant={applicant}
                eventId={eventId}
                open={qrOpen}
                onOpenChange={setQrOpen}
            />

            <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
                <AlertDialogContent className="sm:max-w-md">
                    <AlertDialogHeader>
                        <AlertDialogTitle className="text-base font-semibold">
                            Delete Applicant
                        </AlertDialogTitle>
                        <AlertDialogDescription className="text-xs">
                            Are you sure you want to remove <span className="font-semibold text-foreground">{applicant.name}</span> ({applicant.email})? This will permanently delete their registration and attendance records for this event.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter className="pt-2">
                        <AlertDialogCancel className="text-xs">
                            Cancel
                        </AlertDialogCancel>
                        <AlertDialogAction
                            onClick={handleDelete}
                            disabled={deleteMutation.isPending}
                            className="bg-destructive hover:bg-destructive/90 text-destructive-foreground text-xs gap-1.5"
                        >
                            {
                                deleteMutation.isPending && <Loader2 className="size-3.5 animate-spin" />
                            }
                            <span>
                                {
                                    deleteMutation.isPending ? "Deleting..." : "Delete"
                                }
                            </span>
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}
