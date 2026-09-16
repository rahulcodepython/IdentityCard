"use client";

import * as React from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { FileText } from "lucide-react";

import { getApplicantsColumns } from "@/components/applicants/applicants-columns";
import { ApplicantFiltersBar } from "@/components/applicants/applicant-filters-bar";
import { InfiniteDataTable } from "@/components/generic";
import { Button } from "@/components/ui/button";
import { useBreadcrumbs } from "@/hooks/use-breadcrumbs";
import {
    useApplicantSchemaQuery,
    useApplicantsInfiniteQuery,
} from "@/query-hooks/applicants.api";
import { useEventQuery } from "@/query-hooks/events.api";
import type { ApplicantFilter } from "@/schema/applicants.types";

export default function EventApplicantsPage() {
    const params = useParams();
    const eventId = params?.eventId as string;

    const [search, setSearch] = React.useState("");
    const [filters, setFilters] = React.useState<ApplicantFilter[]>([]);

    const { data: event } = useEventQuery(eventId);

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

    // 1. First API call: Fetch form schema once for column definitions & filter attributes
    const { data: formSchema } = useApplicantSchemaQuery(eventId);

    const formFields = React.useMemo(() => {
        return formSchema?.fields || [];
    }, [formSchema]);

    const columns = React.useMemo(() => {
        return getApplicantsColumns(formFields);
    }, [formFields]);

    // 2. Second API call: Infinite pagination query for applicant data rows
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
                        className="gap-2 text-xs font-medium"
                    >
                        <FileText className="size-3.5" />
                        <span>Registration Form</span>
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
            />
        </div>
    );
}
