"use client";

import * as React from "react";
import type { ColumnDef } from "@tanstack/react-table";

import { DataTable } from "@/components/ui/data-table";
import { DebouncedInput } from "@/components/generic/debounced-input";
import { InfiniteScrollSentinel } from "@/components/generic/infinite-scroll-sentinel";

export interface InfiniteDataTableProps<TData, TValue> {
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
    totalCount?: number;
    isLoading?: boolean;
    isError?: boolean;
    error?: Error | null;
    hasNextPage?: boolean;
    isFetchingNextPage?: boolean;
    fetchNextPage: () => void;
    searchPlaceholder?: string;
    searchValue?: string;
    onSearchChange?: (search: string) => void;
    toolbarActions?: React.ReactNode;
    itemLabel?: string;
}

export function InfiniteDataTable<TData, TValue>({
    columns,
    data,
    totalCount,
    isLoading = false,
    isError = false,
    error = null,
    hasNextPage = false,
    isFetchingNextPage = false,
    fetchNextPage,
    searchPlaceholder = "Search...",
    searchValue,
    onSearchChange,
    toolbarActions,
    itemLabel = "items",
}: InfiniteDataTableProps<TData, TValue>) {
    const count = totalCount ?? data.length;

    return (
        <div className="flex flex-1 flex-col gap-4 w-full max-w-full min-w-0">
            {/* Top Toolbar */}
            {
                (onSearchChange || toolbarActions) && <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    {
                        onSearchChange && <DebouncedInput
                            placeholder={searchPlaceholder}
                            value={searchValue}
                            onChange={onSearchChange}
                        />
                    }
                    {
                        toolbarActions && <div className="flex items-center justify-end">
                            {toolbarActions}
                        </div>
                    }
                </div>
            }

            {/* Error Banner */}
            {
                isError && <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-xs text-destructive">
                    Failed to load {itemLabel}: {error?.message || "Unknown error"}
                </div>
            }

            {/* Data Table */}
            <DataTable
                columns={columns}
                data={data}
                isLoading={isLoading}
                loadingMessage={`Loading ${itemLabel}...`}
                emptyMessage={`No ${itemLabel} found.`}
            />

            {/* Item Counter */}
            <div className="flex items-center justify-between text-xs text-muted-foreground">
                Showing {data.length} of {count} {itemLabel}
            </div>

            {/* Infinite Scroll Sentinel */}
            <InfiniteScrollSentinel
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={fetchNextPage}
                itemCount={data.length}
                itemLabel={itemLabel}
            />
        </div>
    );
}
