"use client";

import * as React from "react";
import Link from "next/link";
import { type ColumnDef } from "@tanstack/react-table";
import { FileEdit, Pencil, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { UpdateFormDialog } from "@/components/forms/update-form-dialog";
import { useDeleteFormMutation } from "@/query-hooks/forms.api";
import type { Form } from "@/schema/forms.types";

function FormRowActions({ form }: { form: Form }) {
    const [editOpen, setEditOpen] = React.useState(false);
    const [deleteOpen, setDeleteOpen] = React.useState(false);
    const deleteMutation = useDeleteFormMutation();

    const handleDelete = async () => {
        try {
            await deleteMutation.mutateAsync(form.id);
            setDeleteOpen(false);
        } catch {
            // Handled by mutation toast
        }
    };

    return (
        <div className="flex items-center justify-end gap-2">
            <Button
                type="button"
                variant="outline"
                className="h-8 gap-2 px-3 text-xs font-medium"
                onClick={() => setEditOpen(true)}
                title="Edit Form Metadata"
            >
                <Pencil className="size-3.5" />
                <span>Edit</span>
            </Button>

            <Button
                type="button"
                variant="outline"
                className="h-8 gap-2 px-3 text-xs font-medium"
                nativeButton={false}
                render={
                    <Link href={`/dashboard/forms/${form.id}/design`} />
                }
            >
                <FileEdit className="size-3.5" />
                <span>Edit Design</span>
            </Button>

            <Button
                type="button"
                variant="destructive"
                className="h-8 gap-2 px-3 text-xs font-medium"
                onClick={() => setDeleteOpen(true)}
                title="Delete Form Template"
            >
                <Trash2 className="size-3.5" />
                <span>Delete</span>
            </Button>

            <UpdateFormDialog
                form={form}
                open={editOpen}
                onOpenChange={setEditOpen}
            />

            <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>Delete Form Template</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to delete <span className="font-semibold text-foreground">&ldquo;{form.name}&rdquo;</span>? This action cannot be undone.
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter className="pt-2">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => setDeleteOpen(false)}
                            disabled={deleteMutation.isPending}
                        >
                            Cancel
                        </Button>
                        <Button
                            type="button"
                            variant="destructive"
                            onClick={handleDelete}
                            disabled={deleteMutation.isPending}
                        >
                            {
                                deleteMutation.isPending ? <span>Deleting...</span> : <span>Delete</span>
                            }
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}

export const formsColumns: ColumnDef<Form>[] = [
    {
        accessorKey: "name",
        header: "Template Name",
        cell: ({ row }) => {
            return (
                <Link
                    className="flex flex-col group"
                    href={`/dashboard/forms/${row.original.id}/design`}
                >
                    <span className="font-semibold text-foreground group-hover:text-primary transition-colors">
                        {row.original.name}
                    </span>
                </Link>
            );
        },
    },
    {
        accessorKey: "fields",
        header: "Fields",
        cell: ({ row }) => {
            const count = row.original.fields?.length ?? 0;
            return (
                <Badge variant="secondary" className="font-normal text-xs">
                    {count} {count === 1 ? "field" : "fields"}
                </Badge>
            );
        },
    },
    {
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => {
            const date = new Date(row.original.created_at);
            return (
                <span className="text-xs text-muted-foreground">
                    {isNaN(date.getTime()) ? "-" : date.toLocaleDateString()}
                </span>
            );
        },
    },
    {
        id: "actions",
        header: () => <div className="text-right">Actions</div>,
        cell: ({ row }) => <FormRowActions form={row.original} />,
    },
];
