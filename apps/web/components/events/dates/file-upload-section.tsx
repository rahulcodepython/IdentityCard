"use client";

import * as React from "react";
import { AlertCircle, CheckCircle2, Eye, Loader2, Upload } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { StagedDatesPreviewDialog } from "@/components/events/dates/staged-dates-preview-dialog";
import type { EventDateItemInput } from "@/schema/event-dates.types";

export interface FileUploadSectionProps {
    eventStartDate: string;
    eventEndDate: string;
    onApplyDates?: (dates: EventDateItemInput[]) => void;
    onOverwrideDates?: (dates: EventDateItemInput[]) => Promise<void>;
    onReplaceAllDates?: (dates: EventDateItemInput[]) => Promise<void>;
    isReplacing?: boolean;
}

type ParseStatus = "idle" | "parsing" | "parsed" | "accepted" | "error";

const MAX_FILE_SIZE_BYTES = 2 * 1024 * 1024; // 2MB

export function FileUploadSection({
    eventStartDate,
    eventEndDate,
    onApplyDates,
    onOverwrideDates,
    onReplaceAllDates,
    isReplacing = false,
}: FileUploadSectionProps) {
    const fileInputRef = React.useRef<HTMLInputElement>(null);

    const [status, setStatus] = React.useState<ParseStatus>("idle");
    const [statusMessage, setStatusMessage] = React.useState<string>("No file chosen");
    const [parsedItems, setParsedItems] = React.useState<EventDateItemInput[]>([]);
    const [previewOpen, setPreviewOpen] = React.useState(false);

    const validateItem = (
        item: unknown,
        idx: number
    ): { valid: true; data: EventDateItemInput } | { valid: false; error: string } => {
        if (!item || typeof item !== "object") {
            return { valid: false, error: `Row ${idx + 1}: Not a valid object` };
        }
        const obj = item as Record<string, unknown>;
        const date = String(obj.date || "").trim();
        const startTime = String(obj.start_time || obj.startTime || "").trim();
        const endTime = String(obj.end_time || obj.endTime || "").trim();

        if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) {
            return { valid: false, error: `Row ${idx + 1}: Invalid date format "${date}" (expected YYYY-MM-DD)` };
        }
        if (!/^\d{2}:\d{2}$/.test(startTime) || !/^\d{2}:\d{2}$/.test(endTime)) {
            return { valid: false, error: `Row ${idx + 1} (${date}): Times must match HH:MM format` };
        }
        if (endTime < startTime) {
            return { valid: false, error: `Row ${idx + 1} (${date}): End time cannot be earlier than start time` };
        }
        if (date < eventStartDate || date > eventEndDate) {
            return {
                valid: false,
                error: `Row ${idx + 1} (${date}): Out of event range [${eventStartDate} to ${eventEndDate}]`,
            };
        }

        return { valid: true, data: { date, start_time: startTime, end_time: endTime } };
    };

    const parseJSON = async (file: File) => {
        if (file.size > MAX_FILE_SIZE_BYTES) {
            setStatus("error");
            setStatusMessage("File size exceeds 2MB limit");
            return;
        }

        setStatus("parsing");
        setStatusMessage("Parsing JSON file...");
        try {
            const text = await file.text();
            const raw = JSON.parse(text);
            if (!Array.isArray(raw)) {
                setStatus("error");
                setStatusMessage("Invalid JSON: Root must be an array of date objects");
                return;
            }

            const results: EventDateItemInput[] = [];
            for (let i = 0; i < raw.length; i++) {
                const res = validateItem(raw[i], i);
                if (!res.valid) {
                    setStatus("error");
                    setStatusMessage(res.error);
                    return;
                }
                results.push(res.data);
            }

            if (results.length === 0) {
                setStatus("error");
                setStatusMessage("JSON file contains no date entries");
                return;
            }

            setParsedItems(results);
            setStatus("accepted");
            setStatusMessage(`Data accepted: ${results.length} valid date entries parsed`);
            toast.success(`Parsed ${results.length} dates from JSON`);
        } catch {
            setStatus("error");
            setStatusMessage("Failed to parse JSON: Syntax error in file");
        }
    };

    const parseCSV = async (file: File) => {
        if (file.size > MAX_FILE_SIZE_BYTES) {
            setStatus("error");
            setStatusMessage("File size exceeds 2MB limit");
            return;
        }

        setStatus("parsing");
        setStatusMessage("Parsing CSV file...");
        try {
            const text = await file.text();
            const lines = text.split(/\r?\n/).map((l) => l.trim()).filter(Boolean);
            if (lines.length < 2) {
                setStatus("error");
                setStatusMessage("CSV must contain a header row and at least one data row");
                return;
            }

            const header = lines[0].toLowerCase().split(",").map((h) => h.trim());
            const dateIdx = header.indexOf("date");
            const startIdx = header.indexOf("start_time") !== -1 ? header.indexOf("start_time") : header.indexOf("starttime");
            const endIdx = header.indexOf("end_time") !== -1 ? header.indexOf("end_time") : header.indexOf("endtime");

            if (dateIdx === -1 || startIdx === -1 || endIdx === -1) {
                setStatus("error");
                setStatusMessage('CSV header must contain "date", "start_time", and "end_time"');
                return;
            }

            const results: EventDateItemInput[] = [];
            for (let i = 1; i < lines.length; i++) {
                const cols = lines[i].split(",").map((c) => c.trim());
                if (cols.length <= Math.max(dateIdx, startIdx, endIdx)) continue;

                const res = validateItem(
                    {
                        date: cols[dateIdx],
                        start_time: cols[startIdx],
                        end_time: cols[endIdx],
                    },
                    i - 1
                );
                if (!res.valid) {
                    setStatus("error");
                    setStatusMessage(res.error);
                    return;
                }
                results.push(res.data);
            }

            if (results.length === 0) {
                setStatus("error");
                setStatusMessage("CSV contains no valid data rows");
                return;
            }

            setParsedItems(results);
            setStatus("accepted");
            setStatusMessage(`Data accepted: ${results.length} valid date entries parsed`);
            toast.success(`Parsed ${results.length} dates from CSV`);
        } catch {
            setStatus("error");
            setStatusMessage("Failed to read CSV file");
        }
    };

    const handleFileSelected = (file: File) => {
        const lowerName = file.name.toLowerCase();
        if (lowerName.endsWith(".json") || file.type === "application/json") {
            parseJSON(file);
        } else if (lowerName.endsWith(".csv") || file.type === "text/csv") {
            parseCSV(file);
        } else {
            setStatus("error");
            setStatusMessage("Invalid file type. Please choose a .json or .csv file.");
        }
    };

    const handleApply = () => {
        if (parsedItems.length === 0 || !onApplyDates) return;
        onApplyDates(parsedItems);
        toast.success(`Imported ${parsedItems.length} dates to calendar staging`);
    };

    const overwrideHandler = onOverwrideDates || onReplaceAllDates;

    const handleOverwride = async () => {
        if (parsedItems.length === 0 || !overwrideHandler) return;
        await overwrideHandler(parsedItems);
    };

    return (
        <div className="flex flex-col gap-3 rounded-md border bg-card p-4">
            <div className="flex flex-col gap-1">
                <span className="text-xs font-semibold text-foreground">Import from File</span>
                <p className="text-[11px] text-muted-foreground">
                    Upload a JSON or CSV file to overwride all dates and schedules (max 2MB).
                </p>
            </div>

            {/* Single input file field accepting both JSON and CSV */}
            <input
                type="file"
                ref={fileInputRef}
                accept=".json,.csv,application/json,text/csv"
                className="hidden"
                onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) {
                        handleFileSelected(file);
                    }
                    e.target.value = "";
                }}
            />

            {/* Single Upload Button */}
            <Button
                type="button"
                variant="outline"
                className="h-9 w-full gap-2 text-xs"
                onClick={() => fileInputRef.current?.click()}
            >
                <Upload className="size-3.5 text-primary" />
                <span>Choose File (JSON or CSV)</span>
            </Button>

            {/* Status Message Display */}
            <div
                className={`flex items-start gap-2  p-2.5 text-xs ${status === "parsing"
                    ? "bg-muted text-muted-foreground"
                    : status === "accepted"
                        ? "border border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400"
                        : status === "error"
                            ? "border border-destructive/20 bg-destructive/10 text-destructive"
                            : "bg-muted/40 text-muted-foreground"
                    }`}
            >
                {
                    status === "parsing" ? <Loader2 className="mt-0.5 size-3.5 animate-spin" />
                        : status === "accepted" ? <CheckCircle2 className="mt-0.5 size-3.5 shrink-0" />
                            : status === "error" ? <AlertCircle className="mt-0.5 size-3.5 shrink-0" />
                                : null
                }
                <span className="leading-tight">{statusMessage}</span>
            </div>

            {/* Actions Bar */}
            {
                status === "accepted" && parsedItems.length > 0 && <div className="flex items-center gap-2 pt-1">
                    <Button
                        type="button"
                        variant="outline"
                        className="h-8 flex-1 gap-1 text-xs"
                        onClick={() => setPreviewOpen(true)}
                    >
                        <Eye className="size-3.5" />
                        <span>Preview ({parsedItems.length})</span>
                    </Button>
                    {
                        overwrideHandler ? <Button
                            type="button"
                            disabled={isReplacing}
                            className="h-8 flex-1 gap-1 text-xs"
                            onClick={handleOverwride}
                        >
                            {
                                isReplacing ? <Loader2 className="size-3.5 animate-spin" />
                                    : <Upload className="size-3.5" />
                            }
                            <span>{isReplacing ? "Overwriding..." : "Overwride All Dates"}</span>
                        </Button> : <Button
                            type="button"
                            className="h-8 flex-1 gap-1 text-xs"
                            onClick={handleApply}
                        >
                            <Upload className="size-3.5" />
                            <span>Apply to Calendar</span>
                        </Button>
                    }
                </div>
            }

            {/* Preview Staged File Items Dialog */}
            <StagedDatesPreviewDialog
                open={previewOpen}
                onOpenChange={setPreviewOpen}
                title="Preview File Import Data"
                description={`Review ${parsedItems.length} dates parsed from the uploaded file before applying.`}
                items={parsedItems.map((p) => ({
                    date: p.date,
                    startTime: p.start_time,
                    endTime: p.end_time,
                    changeType: "added",
                }))}
                actionLabel={overwrideHandler ? "Upload & Overwride All Dates" : "Apply to Calendar"}
                isActionPending={isReplacing}
                onAction={async () => {
                    if (overwrideHandler) {
                        await handleOverwride();
                    } else {
                        handleApply();
                    }
                    setPreviewOpen(false);
                }}
            />
        </div>
    );
}
