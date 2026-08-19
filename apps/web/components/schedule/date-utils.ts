// Local-time-safe date<->"YYYY-MM-DD" conversion for react-day-picker's
// Date objects. Deliberately not `toISOString()` (that's UTC and shifts
// the date near midnight in most timezones).

export function toDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, "0")
  const day = String(d.getDate()).padStart(2, "0")
  return `${y}-${m}-${day}`
}

export function fromDateKey(s: string): Date {
  const [y, m, d] = s.split("-").map(Number)
  return new Date(y, m - 1, d)
}
