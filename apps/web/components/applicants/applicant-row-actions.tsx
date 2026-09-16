"use client";

import * as React from "react";
import { Eye } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { ApplicantItem, FormFieldSummary } from "@/schema/applicants.types";
import { ApplicantDataDialog } from "./applicant-data-dialog";

interface ApplicantRowActionsProps {
    applicant: ApplicantItem;
    formFields?: FormFieldSummary[];
}

export function ApplicantRowActions({
    applicant,
    formFields,
}: ApplicantRowActionsProps) {
    const [detailsOpen, setDetailsOpen] = React.useState(false);

    return (
        <div className="flex items-center justify-end">
            <Button
                type="button"
                variant="outline"
                className="h-8 gap-1 px-2.5 text-xs"
                onClick={() => setDetailsOpen(true)}
            >
                <Eye className="size-3.5" />
                <span>View Data</span>
            </Button>

            <ApplicantDataDialog
                applicant={applicant}
                formFields={formFields}
                open={detailsOpen}
                onOpenChange={setDetailsOpen}
            />
        </div>
    );
}
