"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";

import { Badge } from "@/components/ui/badge";
import type { ApplicantItem, FormFieldSummary } from "@/schema/applicants.types";
import { ApplicantUserIdCell } from "./applicant-user-id-cell";
import { ApplicantRowActions } from "./applicant-row-actions";
import { formatFieldValue } from "./applicants-utils";

export function getApplicantsColumns(
    formFields?: FormFieldSummary[]
): ColumnDef<ApplicantItem>[] {
    const baseColumns: ColumnDef<ApplicantItem>[] = [
        {
            accessorKey: "user_id",
            header: "Applicant ID",
            meta: {
                headerClassName:
                    "sticky left-0 z-30 bg-card w-[140px] min-w-[140px] max-w-[140px]",
                cellClassName:
                    "sticky left-0 z-20 bg-card group-hover/row:bg-muted/50 group-data-[state=selected]/row:bg-muted w-[140px] min-w-[140px] max-w-[140px]",
            },
            cell: ({ row }) => (
                <ApplicantUserIdCell userId={row.original.user_id} />
            ),
        },
        {
            accessorKey: "name",
            header: "Full Name",
            meta: {
                headerClassName:
                    "sticky left-[140px] z-30 bg-card w-[160px] min-w-[160px] max-w-[160px]",
                cellClassName:
                    "sticky left-[140px] z-20 bg-card group-hover/row:bg-muted/50 group-data-[state=selected]/row:bg-muted w-[160px] min-w-[160px] max-w-[160px]",
            },
            cell: ({ row }) => (
                <span
                    className="font-semibold text-foreground text-xs truncate block max-w-[150px]"
                    title={row.original.name}
                >
                    {row.original.name}
                </span>
            ),
        },
        {
            accessorKey: "email",
            header: "Email Address",
            meta: {
                headerClassName:
                    "sticky left-[300px] z-30 bg-card w-[200px] min-w-[200px] max-w-[200px] border-r border-border shadow-[1px_0_0_0_hsl(var(--border))]",
                cellClassName:
                    "sticky left-[300px] z-20 bg-card group-hover/row:bg-muted/50 group-data-[state=selected]/row:bg-muted w-[200px] min-w-[200px] max-w-[200px] border-r border-border shadow-[1px_0_0_0_hsl(var(--border))]",
            },
            cell: ({ row }) => (
                <span
                    className="font-mono text-xs text-muted-foreground truncate block max-w-[190px]"
                    title={row.original.email}
                >
                    {row.original.email}
                </span>
            ),
        },
    ];

    // Filter out standard fields already displayed as direct columns
    const dynamicFields = (formFields || []).filter(
        (f) => f.key !== "name" && f.key !== "email" && f.key !== "user_id"
    );

    const dynamicColumns: ColumnDef<ApplicantItem>[] = dynamicFields.map((field) => ({
        id: `field_${field.key}`,
        header: field.label || field.key,
        meta: {
            headerClassName: "min-w-[160px] max-w-[240px]",
            cellClassName: "min-w-[160px] max-w-[240px]",
        },
        cell: ({ row }) => {
            const val = row.original.data?.[field.key];
            if (val === undefined || val === null || val === "") {
                return <span className="text-xs text-muted-foreground/50">-</span>;
            }

            if (typeof val === "boolean") {
                return (
                    <Badge
                        variant={val ? "default" : "secondary"}
                        className="text-[10px] font-normal"
                    >
                        {val ? "Yes" : "No"}
                    </Badge>
                );
            }

            const display = formatFieldValue(field, val);

            return (
                <span
                    className="text-xs text-foreground truncate max-w-[220px] block"
                    title={display}
                >
                    {display}
                </span>
            );
        },
    }));

    const trailingColumns: ColumnDef<ApplicantItem>[] = [
        {
            accessorKey: "created_at",
            header: "Registered At",
            meta: {
                headerClassName:
                    "sticky right-[100px] z-30 bg-card w-[120px] min-w-[120px] max-w-[120px] border-l border-border shadow-[-1px_0_0_0_hsl(var(--border))]",
                cellClassName:
                    "sticky right-[100px] z-20 bg-card group-hover/row:bg-muted/50 group-data-[state=selected]/row:bg-muted w-[120px] min-w-[120px] max-w-[120px] border-l border-border shadow-[-1px_0_0_0_hsl(var(--border))]",
            },
            cell: ({ row }) => {
                const date = new Date(row.original.created_at);
                return (
                    <span className="text-xs text-muted-foreground whitespace-nowrap">
                        {isNaN(date.getTime()) ? "-" : date.toLocaleDateString()}
                    </span>
                );
            },
        },
        {
            id: "actions",
            header: () => <div className="text-right">Actions</div>,
            meta: {
                headerClassName:
                    "sticky right-0 z-30 bg-card w-[100px] min-w-[100px] max-w-[100px] text-right",
                cellClassName:
                    "sticky right-0 z-20 bg-card group-hover/row:bg-muted/50 group-data-[state=selected]/row:bg-muted w-[100px] min-w-[100px] max-w-[100px] text-right",
            },
            cell: ({ row }) => (
                <ApplicantRowActions
                    applicant={row.original}
                    formFields={formFields}
                />
            ),
        },
    ];

    return [...baseColumns, ...dynamicColumns, ...trailingColumns];
}

export const applicantsColumns: ColumnDef<ApplicantItem>[] = getApplicantsColumns();
