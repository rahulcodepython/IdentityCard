"use client";

import * as React from "react";

import { CreateFormDialog } from "@/components/forms/create-form-dialog";
import { formsColumns } from "@/components/forms/forms-columns";
import { InfiniteDataTable } from "@/components/generic";
import { useBreadcrumbs } from "@/hooks/use-breadcrumbs";
import { useFormsInfiniteQuery } from "@/query-hooks/forms.api";

export default function FormsPage() {
    const [search, setSearch] = React.useState("");

    useBreadcrumbs([
        {
            title: "Dashboard",
            url: "/dashboard",
        },
        {
            title: "Management",
        },
        {
            title: "Forms",
        },
    ]);

    const {
        data,
        isLoading,
        isError,
        error,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useFormsInfiniteQuery({ search });

    const flatForms = React.useMemo(() => {
        return data?.pages.flatMap((page) => page.data) ?? [];
    }, [data]);

    const totalCount = data?.pages[0]?.total ?? flatForms.length;

    return (
        <InfiniteDataTable
            columns={formsColumns}
            data={flatForms}
            totalCount={totalCount}
            isLoading={isLoading}
            isError={isError}
            error={error}
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            fetchNextPage={fetchNextPage}
            searchPlaceholder="Search forms..."
            onSearchChange={setSearch}
            toolbarActions={<CreateFormDialog />}
            itemLabel="forms"
        />
    );
}
