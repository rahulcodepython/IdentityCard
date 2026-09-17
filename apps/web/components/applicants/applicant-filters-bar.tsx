"use client";

import * as React from "react";
import { Filter } from "lucide-react";

import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import type { ApplicantFilter, FormFieldSummary } from "../../schema/applicants.types";
import { ApplicantFilterBadges } from "./applicant-filter-badges";
import { ApplicantFilterForm } from "./applicant-filter-form";

interface ApplicantFiltersBarProps {
    fields: FormFieldSummary[];
    filters: ApplicantFilter[];
    onAddFilter: (filter: ApplicantFilter) => void;
    onRemoveFilter: (index: number) => void;
    onClearFilters: () => void;
}

export function ApplicantFiltersBar({
    fields,
    filters,
    onAddFilter,
    onRemoveFilter,
    onClearFilters,
}: ApplicantFiltersBarProps) {
    const [isOpen, setIsOpen] = React.useState(false);

    return (
        <div className="flex flex-col gap-2.5">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <Button
                    type="button"
                    variant={isOpen ? "secondary" : "outline"}
                    onClick={() => setIsOpen(!isOpen)}
                    className="gap-2"
                >
                    <Filter className="size-3.5 text-primary" />
                    <span>Filter by Fields</span>
                    {filters.length > 0 && (
                        <Badge
                            variant="default"
                            className="ml-1 size-4 rounded-full p-0 text-[10px] flex items-center justify-center font-bold"
                        >
                            {filters.length}
                        </Badge>
                    )}
                </Button>
            </div>

            {/* Filter Builder Form */}
            {isOpen && (
                <ApplicantFilterForm fields={fields} onAddFilter={onAddFilter} />
            )}

            {/* Active Filters Badges */}
            <ApplicantFilterBadges
                filters={filters}
                fields={fields}
                onRemoveFilter={onRemoveFilter}
                onClearFilters={onClearFilters}
            />
        </div>
    );
}
