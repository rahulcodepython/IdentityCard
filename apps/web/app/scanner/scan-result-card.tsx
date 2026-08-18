import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import type { ScanResponse } from "@/lib/validation/scanner"

const directionLabel: Record<ScanResponse["direction"], string> = {
  entry: "Entry recorded",
  exit: "Exit recorded",
  already_completed: "Already checked out today",
}

const statusLabel: Record<NonNullable<ScanResponse["status"]>, string> = {
  early: "Early",
  on_time: "On time",
  late: "Late",
}

// The "simple card popup" the spec calls for: who this is, which event,
// and whether this was their entry or exit scan (first scan of the day =
// entry — see attendance.ScanResponse's Go-side doc comment) with an
// early/on_time/late read against the schedule.
export function ScanResultCard({
  result,
  onDismiss,
}: {
  result: ScanResponse
  onDismiss: () => void
}) {
  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>{directionLabel[result.direction]}</CardTitle>
        <CardDescription>
          {result.event_name} · {result.date}
          {result.status ? ` · ${statusLabel[result.status]}` : ""}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <div>
          <p className="font-medium">{result.person.name}</p>
          <p className="text-sm text-muted-foreground">{result.person.email}</p>
          <p className="text-sm text-muted-foreground">
            {result.person.mobile}
          </p>
        </div>
        <Button onClick={onDismiss} className="mt-2">
          Scan next
        </Button>
      </CardContent>
    </Card>
  )
}
