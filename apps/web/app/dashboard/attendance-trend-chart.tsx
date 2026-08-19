"use client"

import { Area, AreaChart, CartesianGrid, XAxis } from "recharts"

import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart"
import type { DailyTrendPoint } from "@/lib/validation/analytics"

// present/absent colors are reserved status colors used consistently
// across every attendance chart in the app (see the per-event analytics
// page) — never repurposed for anything else.
const chartConfig = {
  present: { label: "Present", color: "var(--primary)" },
  absent: { label: "Absent", color: "var(--muted-foreground)" },
} satisfies ChartConfig

export function AttendanceTrendChart({ data }: { data: DailyTrendPoint[] }) {
  if (data.length === 0) {
    return (
      <div className="flex h-[220px] items-center justify-center text-sm text-muted-foreground">
        No attendance recorded in the last 30 days yet.
      </div>
    )
  }

  return (
    <ChartContainer config={chartConfig} className="h-[220px] w-full">
      <AreaChart data={data} margin={{ left: 0, right: 12 }}>
        <CartesianGrid vertical={false} />
        <XAxis
          dataKey="date"
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          minTickGap={24}
          tickFormatter={(value: string) => value.slice(5)}
        />
        <ChartTooltip content={<ChartTooltipContent indicator="dot" />} />
        <defs>
          <linearGradient id="fillPresent" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-present)" stopOpacity={0.4} />
            <stop offset="95%" stopColor="var(--color-present)" stopOpacity={0.05} />
          </linearGradient>
          <linearGradient id="fillAbsent" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-absent)" stopOpacity={0.35} />
            <stop offset="95%" stopColor="var(--color-absent)" stopOpacity={0.05} />
          </linearGradient>
        </defs>
        <Area
          dataKey="present"
          type="monotone"
          fill="url(#fillPresent)"
          stroke="var(--color-present)"
          stackId="a"
        />
        <Area
          dataKey="absent"
          type="monotone"
          fill="url(#fillAbsent)"
          stroke="var(--color-absent)"
          stackId="a"
        />
      </AreaChart>
    </ChartContainer>
  )
}
