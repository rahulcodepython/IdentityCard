"use client";

import * as React from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import {
    Copy,
    Eye,
    FileText,
    Loader2,
    Mail,
    Phone,
    QrCode,
    Trash2,
    UserPlus,
} from "lucide-react";
import { toast } from "sonner";

import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogBody,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "../../../../../components/ui/alert-dialog";
import {
    ContextMenuItem,
    ContextMenuSeparator,
} from "../../../../../components/ui/context-menu";
import { ApplicantDataDialog } from "../../../../../components/applicants/applicant-data-dialog";
import { ApplicantFiltersBar } from "../../../../../components/applicants/applicant-filters-bar";
import { ApplicantQrDialog } from "../../../../../components/applicants/applicant-qr-dialog";
import { CreateApplicantDialog } from "../../../../../components/applicants/create-applicant-dialog";
import { getApplicantsColumns } from "../../../../../components/applicants/applicants-columns";
import { InfiniteDataTable } from "../../../../../components/generic";
import { Button } from "../../../../../components/ui/button";
import { useBreadcrumbs } from "../../../../../hooks/use-breadcrumbs";
import {
    useApplicantSchemaQuery,
    useApplicantsInfiniteQuery,
    useDeleteApplicantMutation,
} from "../../../../../query-hooks/applicants.api";
import { useEventQuery } from "../../../../../query-hooks/events.api";
import type { ApplicantFilter, ApplicantItem } from "../../../../../schema/applicants.types";

export default function EventApplicantsPage() {
    const params = useParams();
    const eventId = params?.eventId as string;

    const [search, setSearch] = React.useState("");
    const [filters, setFilters] = React.useState<ApplicantFilter[]>([]);
    const [createOpen, setCreateOpen] = React.useState(false);
    const [viewDataApplicant, setViewDataApplicant] = React.useState<ApplicantItem | null>(null);
    const [viewQrApplicant, setViewQrApplicant] = React.useState<ApplicantItem | null>(null);
    const [deleteApplicant, setDeleteApplicant] = React.useState<ApplicantItem | null>(null);

    const { data: event } = useEventQuery(eventId);
    const deleteMutation = useDeleteApplicantMutation(eventId);

    useBreadcrumbs([
        {
            title: "Dashboard",
            url: "/dashboard",
        },
        {
            title: "Events",
            url: "/dashboard/events",
        },
        {
            title: event?.name || "Event",
            url: `/dashboard/events/${eventId}`,
        },
        {
            title: "Applicants",
        },
    ]);

    // 1. Fetch form schema once for column definitions & filter attributes
    const { data: formSchema } = useApplicantSchemaQuery(eventId);

    const formFields = React.useMemo(() => {
        return formSchema?.fields || [];
    }, [formSchema]);

    const columns = React.useMemo(() => {
        return getApplicantsColumns(formFields, eventId);
    }, [formFields, eventId]);

    // 2. Infinite pagination query for applicant data rows
    const {
        data,
        isLoading,
        isError,
        error,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useApplicantsInfiniteQuery(eventId, { search, filters });

    const handleAddFilter = React.useCallback((newFilter: ApplicantFilter) => {
        setFilters((prev) => [...prev, newFilter]);
    }, []);

    const handleRemoveFilter = React.useCallback((index: number) => {
        setFilters((prev) => prev.filter((_, i) => i !== index));
    }, []);

    const handleClearFilters = React.useCallback(() => {
        setFilters([]);
    }, []);

    const flatApplicants = React.useMemo(() => {
        return data?.pages.flatMap((page) => page.data) ?? [];
    }, [data]);

    const totalCount = data?.pages[0]?.total ?? flatApplicants.length;

    const handleDelete = async () => {
        if (!deleteApplicant) return;
        try {
            await deleteMutation.mutateAsync(deleteApplicant.user_id);
            setDeleteApplicant(null);
        } catch {
            // Handled by mutation toast
        }
    };

    const renderRowContextMenu = (row: { original: ApplicantItem }) => {
        const applicant = row.original;
        return (
            <>
                <ContextMenuItem
                    onClick={() => setViewDataApplicant(applicant)}
                    className="gap-2.5 font-medium"
                >
                    <Eye className="size-3.5 text-primary" />
                    <span>View Form Data</span>
                </ContextMenuItem>

                <ContextMenuItem
                    onClick={() => setViewQrApplicant(applicant)}
                    className="gap-2.5"
                >
                    <QrCode className="size-3.5 text-primary" />
                    <span>View QR Code</span>
                </ContextMenuItem>

                <ContextMenuSeparator />

                <ContextMenuItem
                    onClick={() => {
                        navigator.clipboard.writeText(applicant.user_id);
                        toast.success("Applicant ID copied");
                    }}
                    className="gap-2.5"
                >
                    <Copy className="size-3.5" />
                    <span>Copy Applicant ID</span>
                </ContextMenuItem>

                <ContextMenuItem
                    onClick={() => {
                        navigator.clipboard.writeText(applicant.email);
                        toast.success("Email copied");
                    }}
                    className="gap-2.5"
                >
                    <Mail className="size-3.5" />
                    <span>Copy Email Address</span>
                </ContextMenuItem>

                {applicant.phone && (
                    <ContextMenuItem
                        onClick={() => {
                            navigator.clipboard.writeText(applicant.phone);
                            toast.success("Mobile number copied");
                        }}
                        className="gap-2.5"
                    >
                        <Phone className="size-3.5" />
                        <span>Copy Mobile Number</span>
                    </ContextMenuItem>
                )}

                <ContextMenuSeparator />

                <ContextMenuItem
                    onClick={() => setDeleteApplicant(applicant)}
                    variant="destructive"
                    className="gap-2.5"
                >
                    <Trash2 className="size-3.5" />
                    <span>Delete Applicant</span>
                </ContextMenuItem>
            </>
        );
    };

    return (
        <div className="flex flex-col gap-4 w-full max-w-full min-w-0">
            <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between px-1">
                <div>
                    <h1 className="text-lg font-bold text-foreground">
                        Registered Applicants
                    </h1>
                    <p className="text-xs text-muted-foreground">
                        All verified attendees registered for {event?.name || "this event"}.
                    </p>
                </div>

                <div className="flex items-center gap-2">
                    <Button
                        type="button"
                        variant="outline"
                        nativeButton={false}
                        render={<Link href={`/dashboard/events/${eventId}/form`} />}
                        className="gap-2 font-medium"
                    >
                        <FileText className="size-4" />
                        <span>Registration Form</span>
                    </Button>

                    <Button
                        type="button"
                        variant="default"
                        onClick={() => setCreateOpen(true)}
                        className="gap-2 font-semibold"
                    >
                        <UserPlus className="size-4" />
                        <span>Add Applicant</span>
                    </Button>
                </div>
            </div>

            <ApplicantFiltersBar
                fields={formFields}
                filters={filters}
                onAddFilter={handleAddFilter}
                onRemoveFilter={handleRemoveFilter}
                onClearFilters={handleClearFilters}
            />

            <InfiniteDataTable
                columns={columns}
                data={flatApplicants}
                totalCount={totalCount}
                isLoading={isLoading}
                isError={isError}
                error={error}
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={fetchNextPage}
                searchPlaceholder="Search by name, email, or applicant ID..."
                onSearchChange={setSearch}
                itemLabel="applicants"
                renderRowContextMenu={renderRowContextMenu}
            />

            <CreateApplicantDialog
                eventId={eventId}
                formFields={formFields}
                open={createOpen}
                onOpenChange={setCreateOpen}
            />

            {viewDataApplicant && (
                <ApplicantDataDialog
                    applicant={viewDataApplicant}
                    formFields={formFields}
                    open={!!viewDataApplicant}
                    onOpenChange={(open) => {
                        if (!open) setViewDataApplicant(null);
                    }}
                />
            )}

            {viewQrApplicant && (
                <ApplicantQrDialog
                    applicant={viewQrApplicant}
                    eventId={eventId}
                    open={!!viewQrApplicant}
                    onOpenChange={(open) => {
                        if (!open) setViewQrApplicant(null);
                    }}
                />
            )}

            <AlertDialog
                open={!!deleteApplicant}
                onOpenChange={(open) => {
                    if (!open) setDeleteApplicant(null);
                }}
            >
                <AlertDialogContent className="sm:max-w-lg">
                    <AlertDialogHeader>
                        <AlertDialogTitle className="text-destructive">
                            Delete Applicant
                        </AlertDialogTitle>
                    </AlertDialogHeader>
                    <AlertDialogBody>
                        <AlertDialogDescription>
                            Are you sure you want to remove{" "}
                            <span className="font-semibold text-foreground">
                                {deleteApplicant?.name}
                            </span>{" "}
                            ({deleteApplicant?.email})? This will permanently delete their
                            registration and attendance records for this event.
                        </AlertDialogDescription>
                    </AlertDialogBody>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={deleteMutation.isPending}>
                            Cancel
                        </AlertDialogCancel>
                        <AlertDialogAction
                            onClick={handleDelete}
                            disabled={deleteMutation.isPending}
                            className="bg-destructive hover:bg-destructive/90 text-destructive-foreground gap-1.5"
                        >
                            {deleteMutation.isPending && (
                                <Loader2 className="size-4 animate-spin" />
                            )}
                            <span>
                                {deleteMutation.isPending ? "Deleting..." : "Delete"}
                            </span>
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}
