"use client"

import {
    type ColumnDef,
    type ColumnFiltersState,
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
    type SortingState,
    useReactTable,
    type VisibilityState,
} from "@tanstack/react-table"
import {
    RiArrowDownSLine,
    RiArrowUpSLine,
    RiArrowDownLine,
    RiExpandUpDownLine,
} from "@remixicon/react"
import { ChevronDown, Columns3 } from "lucide-react"
import { useEffect, useState } from "react"

import { Button } from "@/components/ui/button"
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"

interface DataTableProps<TData, TValue> {
    columns: ColumnDef<TData, TValue>[]
    data: TData[]
    searchPlaceholder?: string
    emptyMessage?: string
    pageSize?: number
    extraActions?: React.ReactNode
    getRowClassName?: (row: any) => string
}

export function DataTable<TData, TValue>({
    columns,
    data,
    searchPlaceholder = "Search...",
    emptyMessage = "No results.",
    pageSize = 10,
    extraActions,
    getRowClassName,
}: DataTableProps<TData, TValue>) {
    const [sorting, setSorting] = useState<SortingState>([])
    const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
    const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
    const [globalFilter, setGlobalFilter] = useState("")
    const [visibleCount, setVisibleCount] = useState(pageSize)

    const table = useReactTable({
        data,
        columns,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
        getFilteredRowModel: getFilteredRowModel(),
        onSortingChange: setSorting,
        onColumnFiltersChange: setColumnFilters,
        onColumnVisibilityChange: setColumnVisibility,
        onGlobalFilterChange: setGlobalFilter,
        state: { sorting, columnFilters, columnVisibility, globalFilter },
    })

    const rows = table.getRowModel().rows

    useEffect(() => {
        setVisibleCount(pageSize)
    }, [globalFilter, columnFilters, pageSize])

    const visibleRows = rows.slice(0, visibleCount)
    const hasMore = visibleCount < rows.length

    return (
        <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-2.5 sm:flex-row sm:items-center sm:justify-between">
                <Input
                    placeholder={searchPlaceholder}
                    value={globalFilter}
                    onChange={(e) => table.setGlobalFilter(e.target.value)}
                    className="max-w-xs"
                />
                <div className="flex flex-wrap items-center gap-2">
                    {extraActions}
                    <DropdownMenu>
                        <DropdownMenuTrigger
                            render={
                                <Button variant="outline" size="sm">
                                    <Columns3 className="size-4" />
                                    Columns
                                    <ChevronDown data-icon="inline-end" />
                                </Button>
                            }
                        />
                        <DropdownMenuContent align="end">
                            {table
                                .getAllColumns()
                                .filter((column) => column.getCanHide())
                                .map((column) => (
                                    <DropdownMenuCheckboxItem
                                        key={column.id}
                                        checked={column.getIsVisible()}
                                        onCheckedChange={(checked) =>
                                            column.toggleVisibility(!!checked)
                                        }
                                        onSelect={(e) => e.preventDefault()}
                                        className="capitalize"
                                    >
                                        {column.id}
                                    </DropdownMenuCheckboxItem>
                                ))}
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </div>

            <div className="rounded-xl border overflow-hidden">
                <Table>
                    <TableHeader>
                        {table.getHeaderGroups().map((headerGroup) => (
                            <TableRow key={headerGroup.id}>
                                {headerGroup.headers.map((header) => {
                                    const sortable = header.column.getCanSort()
                                    return (
                                        <TableHead key={header.id}>
                                            {header.isPlaceholder ? null : sortable ? (
                                                <button
                                                    type="button"
                                                    className="flex items-center gap-1 hover:text-foreground"
                                                    onClick={header.column.getToggleSortingHandler()}
                                                >
                                                    {flexRender(
                                                        header.column.columnDef.header,
                                                        header.getContext()
                                                    )}
                                                    {{
                                                        asc: <RiArrowUpSLine className="size-3.5" />,
                                                        desc: <RiArrowDownSLine className="size-3.5" />,
                                                    }[header.column.getIsSorted() as string] ?? (
                                                            <RiExpandUpDownLine className="size-3.5 text-muted-foreground" />
                                                        )}
                                                </button>
                                            ) : (
                                                flexRender(
                                                    header.column.columnDef.header,
                                                    header.getContext()
                                                )
                                            )}
                                        </TableHead>
                                    )
                                })}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody>
                        {visibleRows.length ? (
                            visibleRows.map((row) => (
                                <TableRow
                                    key={row.id}
                                    className={getRowClassName ? getRowClassName(row) : undefined}
                                >
                                    {row.getVisibleCells().map((cell) => (
                                        <TableCell key={cell.id}>
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </TableCell>
                                    ))}
                                </TableRow>
                            ))
                        ) : (
                            <TableRow>
                                <TableCell
                                    colSpan={table.getVisibleFlatColumns().length || 1}
                                    className="h-24 text-center text-muted-foreground"
                                >
                                    {emptyMessage}
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>

            <div className="flex items-center justify-between">
                <p className="text-xs text-muted-foreground">
                    {rows.length === 0
                        ? null
                        : `Showing ${visibleRows.length} of ${rows.length}`}
                </p>
                {hasMore && (
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setVisibleCount((c) => c + pageSize)}
                    >
                        Load more
                        <RiArrowDownLine data-icon="inline-end" />
                    </Button>
                )}
            </div>
        </div>
    )
}
