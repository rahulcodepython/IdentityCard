"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { useImportDaysMutation } from "@/query-hooks/events.api";
import type { DayImportSummary } from "@/schema/events.types";

type ImportState = { summary: DayImportSummary } | { error: string } | null;

function ImportResult({ state }: { state: ImportState }) {
    if (!state) return null;
    if ("error" in state) {
        return <p className="text-sm text-destructive">{state.error}</p>;
    }
    return (
        <div className="flex flex-col gap-2 rounded-lg border p-4 text-sm">
            <p>
                Imported {state.summary.imported}, skipped {state.summary.skipped}.
            </p>
            {state.summary.errors.length > 0 && (
                <ul className="list-disc pl-4 text-muted-foreground">
                    {state.summary.errors.map((rowError, i) => (
                        <li key={i}>
                            Row {rowError.row}: {rowError.message}
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
}

export function ImportForm({ eventId }: { eventId: string }) {
    const importMutation = useImportDaysMutation(eventId);
    const [daysState, setDaysState] = useState<ImportState>(null);

    async function submitDays(e: React.FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        setDaysState(null);

        const result = await importMutation.execute(formData);
        if (result) {
            setDaysState({ summary: result });
        } else if (importMutation.error) {
            setDaysState({ error: importMutation.error.message || "Something went wrong." });
        }
    }

    return (
        <div className="flex flex-col gap-8">
            <div className="flex flex-col gap-4">
                <h2 className="font-heading text-lg font-medium">Import days</h2>
                <form onSubmit={submitDays} className="flex max-w-lg flex-col gap-4">
                    <div className="flex flex-col gap-1.5">
                        <Label htmlFor="days-file">CSV file</Label>
                        <input
                            id="days-file"
                            name="file"
                            type="file"
                            accept=".csv,text/csv"
                            required
                            className="text-sm"
                        />
                        <p className="text-xs text-muted-foreground">
                            Required columns: date, entry_time, exit_time. Replaces the entire day list.
                        </p>
                    </div>
                    <Button type="submit" disabled={importMutation.isPending} className="self-start">
                        {importMutation.isPending ? "Importing…" : "Import days"}
                    </Button>
                </form>
                <ImportResult state={daysState} />
            </div>
        </div>
    );
}