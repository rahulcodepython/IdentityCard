"use client"

import type { ScheduleMode } from "@/lib/validation/events"

const MODES: { value: ScheduleMode; label: string; description: string }[] = [
  { value: "flash", label: "Flash", description: "A single static day" },
  { value: "fixed_range", label: "Fixed range", description: "Every day in a date range, one time window" },
  { value: "selective", label: "Selective dates", description: "Pick specific dates, each its own time" },
  { value: "recurring", label: "Recurring", description: "A weekly pattern, optionally open-ended" },
]

// `allowFlash` gates the Flash option on the org's current plan — only a
// Flash-plan subscription can fund a schedule_mode='flash' event (see
// events.Service.checkEventLimit's flash cross-check).
export function ModeSelect({
  value,
  onChange,
  allowFlash,
}: {
  value: ScheduleMode
  onChange: (mode: ScheduleMode) => void
  allowFlash: boolean
}) {
  const modes = MODES.filter((m) => m.value !== "flash" || allowFlash)

  return (
    <div className="grid grid-cols-2 gap-2">
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
  )
}
