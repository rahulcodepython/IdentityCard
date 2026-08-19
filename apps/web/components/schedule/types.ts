export type DayEntry = { date: string; entry_time: string; exit_time: string }
export type WeekdayEntry = { weekday: number; entry_time: string; exit_time: string }
export type RecurrenceValue = {
  starts_on: string
  ends_on: string | null
  weekdays: WeekdayEntry[]
}

// 0=Monday..6=Sunday — matches apps/server/internal/modules/events'
// RecurrenceWeekdayInput.
export const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
