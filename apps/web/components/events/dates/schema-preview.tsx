"use client";

import * as React from "react";
import { Check, Copy } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import type { EventDate } from "@/schema/event-dates.types";

const JSON_SCHEMA_EXAMPLE = `[
    {
        "date": "2026-10-15",
        "start_time": "09:00",
        "end_time": "18:00"
    },
    {
        "date": "2026-10-16",
        "start_time": "10:00",
        "end_time": "17:30"
    }
]`;

const CSV_SCHEMA_EXAMPLE = `date,start_time,end_time
2026-10-15,09:00,18:00
2026-10-16,10:00,17:30`;

export interface SchemaPreviewProps {
    currentDates: EventDate[];
}

export function SchemaPreview({ currentDates }: SchemaPreviewProps) {
    const [viewFormat, setViewFormat] = React.useState<"json" | "csv">("json");
    const [sourceMode, setSourceMode] = React.useState<"template" | "saved">("template");
    const [hasCopied, setHasCopied] = React.useState(false);

    const containerRef = React.useRef<HTMLDivElement>(null);
    const [maxLines, setMaxLines] = React.useState(8);

    // Compute complete un-truncated content
    const activeFullContent = React.useMemo(() => {
        if (sourceMode === "template") {
            return viewFormat === "json" ? JSON_SCHEMA_EXAMPLE : CSV_SCHEMA_EXAMPLE;
        }

        if (currentDates.length === 0) {
            return viewFormat === "json" ? "[]" : "date,start_time,end_time";
        }

        if (viewFormat === "json") {
            const formatted = currentDates.map((d) => ({
                date: d.date,
                start_time: d.start_time,
                end_time: d.end_time,
            }));
            return JSON.stringify(formatted, null, 4);
        }

        const lines = ["date,start_time,end_time"];
        for (const d of currentDates) {
            lines.push(`${d.date},${d.start_time},${d.end_time}`);
        }
        return lines.join("\n");
    }, [sourceMode, viewFormat, currentDates]);

    // Measure container height to calculate lines that fit without scrollbars
    React.useEffect(() => {
        if (!containerRef.current) return;
        const el = containerRef.current;

        const updateLines = () => {
            const h = el.clientHeight;
            if (h > 0) {
                // 18px line height, 24px vertical padding
                const fit = Math.max(4, Math.floor((h - 24) / 18));
                setMaxLines(fit);
            }
        };

        updateLines();
        const observer = new ResizeObserver(updateLines);
        observer.observe(el);
        return () => observer.disconnect();
    }, []);

    // Format content with middle ellipsis ("....") when lines exceed permitted box height
    const displayContent = React.useMemo(() => {
        const lines = activeFullContent.split("\n");
        if (lines.length <= maxLines) {
            return activeFullContent;
        }

        const available = maxLines - 1; // reserve 1 line for ellipsis
        const headCount = Math.max(1, Math.floor(available / 2));
        const tailCount = Math.max(1, available - headCount);

        return [
            ...lines.slice(0, headCount),
            "    ....",
            ...lines.slice(lines.length - tailCount),
        ].join("\n");
    }, [activeFullContent, maxLines]);

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(activeFullContent);
            setHasCopied(true);
            toast.success("Copied full content to clipboard");
            setTimeout(() => setHasCopied(false), 2000);
        } catch {
            toast.error("Failed to copy content");
        }
    };

    return (
        <div className="flex flex-col gap-3 h-full">
            {/* Header controls */}
            <div className="flex flex-wrap items-center justify-between gap-2">
                {/* Source toggle: Template Schema vs Current Saved */}
                <div className="flex items-center gap-1 rounded-lg border bg-muted/40 p-0.5 text-xs">
                    <button
                        type="button"
                        onClick={() => setSourceMode("template")}
                        className={`rounded-md px-2 py-1 font-medium transition-colors ${sourceMode === "template"
                            ? "bg-background text-foreground shadow-xs"
                            : "text-muted-foreground hover:text-foreground"
                            }`}
                    >
                        Schema Template
                    </button>
                    <button
                        type="button"
                        onClick={() => setSourceMode("saved")}
                        className={`rounded-md px-2 py-1 font-medium transition-colors ${sourceMode === "saved"
                            ? "bg-background text-foreground shadow-xs"
                            : "text-muted-foreground hover:text-foreground"
                            }`}
                    >
                        Saved Data ({currentDates.length})
                    </button>
                </div>

                {/* Format toggle: JSON vs CSV & Copy Button */}
                <div className="flex items-center gap-1.5">
                    <div className="flex items-center gap-1 rounded-lg border bg-muted/40 p-0.5 text-xs">
                        <button
                            type="button"
                            onClick={() => setViewFormat("json")}
                            className={`rounded-md px-2 py-1 font-medium transition-colors ${viewFormat === "json"
                                ? "bg-background text-foreground shadow-xs"
                                : "text-muted-foreground hover:text-foreground"
                                }`}
                        >
                            JSON
                        </button>
                        <button
                            type="button"
                            onClick={() => setViewFormat("csv")}
                            className={`rounded-md px-2 py-1 font-medium transition-colors ${viewFormat === "csv"
                                ? "bg-background text-foreground shadow-xs"
                                : "text-muted-foreground hover:text-foreground"
                                }`}
                        >
                            CSV
                        </button>
                    </div>

                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-7 text-muted-foreground hover:text-foreground"
                        onClick={handleCopy}
                        title="Copy full un-truncated code"
                    >
                        {hasCopied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
                    </Button>
                </div>
            </div>

            {/* Non-scrolling, truncated code view */}
            <div ref={containerRef} className="relative h-full rounded-lg border bg-muted/30 p-3 overflow-hidden font-mono text-[11px] leading-4.5 text-muted-foreground">
                <pre className="overflow-hidden whitespace-pre">
                    {displayContent}
                </pre>
            </div>
        </div>
    );
}
