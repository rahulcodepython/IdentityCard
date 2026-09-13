import { Button } from "@/components/ui/button";
import type { RosterEntry } from "@/schema/attendance.types";

const statusLabel: Record<string, string> = {
    early: "Early",
    on_time: "On time",
    late: "Late",
};

export function RosterTable({
    roster,
    exportHref,
}: {
    roster: RosterEntry[];
    exportHref: string;
}) {
    return (
        <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between">
                <h2 className="font-heading text-lg font-medium">Roster</h2>
                <Button variant="outline" render={<a href={exportHref} />}>
                    Export CSV
                </Button>
            </div>

            {roster.length === 0 ? (
                <p className="text-sm text-muted-foreground">No matching records.</p>
            ) : (
                <div className="overflow-x-auto rounded-lg border">
                    <table className="w-full text-sm">
                        <thead className="border-b bg-muted/50 text-left text-muted-foreground">
                            <tr>
                                <th className="p-2 font-medium">Name</th>
                                <th className="p-2 font-medium">Date</th>
                                <th className="p-2 font-medium">Attended</th>
                                <th className="p-2 font-medium">Entry</th>
                                <th className="p-2 font-medium">Exit</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y">
                            {roster.map((entry) => (
                                <tr key={`${entry.person_id}-${entry.date}`}>
                                    <td className="p-2">{entry.person_name}</td>
                                    <td className="p-2 text-muted-foreground">{entry.date}</td>
                                    <td className="p-2">{entry.attended ? "Yes" : "No"}</td>
                                    <td className="p-2 text-muted-foreground">
                                        {entry.entry_at
                                            ? `${new Date(entry.entry_at).toLocaleTimeString()} · ${statusLabel[entry.entry_status ?? ""] ?? "—"}`
                                            : "—"}
                                    </td>
                                    <td className="p-2 text-muted-foreground">
                                        {entry.exit_at
                                            ? `${new Date(entry.exit_at).toLocaleTimeString()} · ${statusLabel[entry.exit_status ?? ""] ?? "—"}`
                                            : "—"}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
}
