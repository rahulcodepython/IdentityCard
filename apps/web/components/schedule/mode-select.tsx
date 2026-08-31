"use client";

import type { EventType } from "@/schema/events.types";

const MODES: { value: EventType; label: string; description: string }[] = [
    { value: "flash", label: "Flash", description: "A single static day" },
    { value: "standard", label: "Standard", description: "Multi-day scheduled event" },
    { value: "grouped", label: "Grouped", description: "Date range with sub-events" },
];

export function ModeSelect({
    value,
    onChange,
    allowFlash,
}: {
    value: EventType;
    onChange: (mode: EventType) => void;
    allowFlash: boolean;
}) {
    const modes = MODES.filter((m) => m.value !== "flash" || allowFlash);

    return (
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
            {modes.map((m) => (
                <button
                    key={m.value}
                    type="button"
                    onClick={() => onChange(m.value)}
                    className={
                        "flex flex-col items-start gap-0.5 rounded-lg border p-3 text-left text-sm transition-colors " +
                        (value === m.value ? "border-primary ring-1 ring-primary" : "hover:bg-muted")
                    }
                >
                    <span className="font-medium">{m.label}</span>
                    <span className="text-xs text-muted-foreground">{m.description}</span>
                </button>
            ))}
        </div>
    );
}
