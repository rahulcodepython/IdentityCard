"use client";

import {
    type BreadcrumbItem,
    useSetBreadcrumbLabel,
    useSetBreadcrumbs,
} from "@/store/breadcrumb.store";

export function BreadcrumbSetter({
    id,
    label,
    items,
}: {
    id?: string;
    label?: string;
    items?: BreadcrumbItem[];
}) {
    useSetBreadcrumbLabel(id, label);
    useSetBreadcrumbs(items ?? []);
    return null;
}
