"use client"

import { useRef } from "react"

import { Button } from "@/components/ui/button"
import { downloadSvgAsPng } from "@/lib/svg-download"

export type ChartSeries = { key: string; label: string; color: string }

// A small hand-rolled SVG grouped bar chart — deliberately not a charting
// library, since the whole app has exactly this one chart type used four
// times. One axis (counts), a fixed color per series never reassigned
// between charts (see the STATUS_SERIES/attendance-series constants in
// page.tsx), a legend whenever there's more than one series, and a
// native <title> per bar as a minimal hover — the real accessible
// fallback is the data table rendered alongside every chart on this page,
// not a bespoke tooltip layer.
export function GroupedBarChart({
  categories,
  series,
  data,
  filename,
}: {
  categories: string[]
  series: ChartSeries[]
  data: Record<string, number>[]
  filename: string
}) {
  const svgRef = useRef<SVGSVGElement>(null)

  const width = 640
  const height = 260
  const paddingLeft = 32
  const paddingBottom = 28
  const paddingTop = 12
  const paddingRight = 12
  const plotWidth = width - paddingLeft - paddingRight
  const plotHeight = height - paddingTop - paddingBottom

  const maxValue = Math.max(
    1,
    ...data.flatMap((row) => series.map((s) => row[s.key] ?? 0))
  )
  const niceMax = Math.max(1, Math.ceil(maxValue / 5) * 5)

  const groupCount = Math.max(categories.length, 1)
  const groupWidth = plotWidth / groupCount
  const barGap = 3
  const barWidth = Math.max(
    3,
    (groupWidth - barGap * (series.length + 1)) / series.length
  )

  const yFor = (value: number) =>
    paddingTop + plotHeight - (value / niceMax) * plotHeight

  if (categories.length === 0) {
    return <p className="text-sm text-muted-foreground">No data yet.</p>
  }

  return (
    <div className="flex flex-col gap-2">
      {series.length > 1 && (
        <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
          {series.map((s) => (
            <span key={s.key} className="flex items-center gap-1.5">
              <span
                className="inline-block size-2.5 rounded-full"
                style={{ backgroundColor: s.color }}
              />
              {s.label}
            </span>
          ))}
        </div>
      )}

      <svg
        ref={svgRef}
        width={width}
        height={height}
        viewBox={`0 0 ${width} ${height}`}
        className="w-full"
        role="img"
        aria-label="Bar chart"
      >
        {[0, 0.25, 0.5, 0.75, 1].map((f) => {
          const y = paddingTop + plotHeight * (1 - f)
          return (
            <line
              key={f}
              x1={paddingLeft}
              x2={width - paddingRight}
              y1={y}
              y2={y}
              stroke="var(--border)"
              strokeWidth={1}
            />
          )
        })}

        {categories.map((category, i) => {
          const groupX = paddingLeft + i * groupWidth
          const row = data[i] ?? {}
          return (
            <g key={category}>
              {series.map((s, si) => {
                const value = row[s.key] ?? 0
                const barX = groupX + barGap + si * (barWidth + barGap)
                const barY = yFor(value)
                const barHeight = Math.max(0, paddingTop + plotHeight - barY)
                return (
                  <rect
                    key={s.key}
                    x={barX}
                    y={barY}
                    width={barWidth}
                    height={barHeight}
                    rx={2}
                    fill={s.color}
                  >
                    <title>
                      {category} · {s.label}: {value}
                    </title>
                  </rect>
                )
              })}
              <text
                x={groupX + groupWidth / 2}
                y={height - paddingBottom + 14}
                textAnchor="middle"
                fontSize={9}
                fill="var(--muted-foreground)"
              >
                {category.length > 10 ? category.slice(5) : category}
              </text>
            </g>
          )
        })}

        {[0, 0.5, 1].map((f) => (
          <text
            key={f}
            x={paddingLeft - 6}
            y={paddingTop + plotHeight * (1 - f) + 3}
            textAnchor="end"
            fontSize={9}
            fill="var(--muted-foreground)"
          >
            {Math.round(niceMax * f)}
          </text>
        ))}
      </svg>

      <Button
        variant="outline"
        size="sm"
        className="self-end"
        onClick={() =>
          svgRef.current && downloadSvgAsPng(svgRef.current, filename)
        }
      >
        Download PNG
      </Button>
    </div>
  )
}
