"use client"

import { RiArrowRightLine, RiBuilding2Line } from "@remixicon/react"

import { Badge } from "@/components/ui/badge"
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import useOrganization from "@/hooks/use-organization"

export default function DashboardIndexPage() {
    const { organizations, handleSelectOrg } = useOrganization(null)

    return (
        <main className="flex-1">
            <div className="mx-auto max-w-6xl px-4 py-12 sm:px-6 lg:px-8 flex flex-col gap-14">
                {/* Organization Selection Header */}
                <div className="flex flex-col items-center text-center gap-3 max-w-2xl mx-auto">
                    <Badge
                        variant="secondary"
                        className="px-3 py-1 text-xs font-medium"
                    >
                        Select Workspace
                    </Badge>
                    <h1 className="font-heading text-3xl font-bold tracking-tight sm:text-4xl text-foreground">
                        Choose an Organization
                    </h1>
                    <p className="text-muted-foreground text-sm sm:text-base">
                        Select an existing organization workspace to manage your
                        events, badges, and scanner devices.
                    </p>
                </div>

                {/* Organization Cards */}
                {organizations && organizations.length > 0 ? (
                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                        {organizations.map((org) => (
                            <Card
                                key={org.id}
                                className={
                                    "group relative cursor-pointer border transition-all duration-200 hover:border-primary hover:shadow-md"
                                }
                                onClick={() => handleSelectOrg(org.slug)}
                            >
                                <CardHeader className="flex flex-row items-start justify-between pb-3">
                                    <div className="flex size-11 items-center justify-center rounded-xl bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
                                        <RiBuilding2Line className="size-5" />
                                    </div>
                                    <Badge
                                        variant="outline"
                                        className="text-[11px] text-muted-foreground font-mono"
                                    >
                                        /{org.slug}
                                    </Badge>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="space-y-1">
                                        <CardTitle className="text-lg font-bold group-hover:text-primary transition-colors">
                                            {org.name}
                                        </CardTitle>
                                        <CardDescription className="line-clamp-1 text-xs">
                                            Organization Workspace
                                        </CardDescription>
                                    </div>

                                    <div className="flex items-center justify-between pt-2 border-t text-xs text-muted-foreground">
                                        <span>Click to enter</span>
                                        <div className="flex items-center justify-center gap-1 font-medium text-foreground group-hover:text-primary transition-colors">
                                            <span>Open</span>
                                            <RiArrowRightLine className="size-3.5 transition-transform group-hover:translate-x-0.5" />
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                ) : (
                    <Card className="max-w-md mx-auto text-center py-8">
                        <CardContent className="space-y-4">
                            <div className="flex size-12 mx-auto items-center justify-center rounded-2xl bg-muted text-muted-foreground">
                                <RiBuilding2Line className="size-6" />
                            </div>
                            <div className="space-y-1.5">
                                <CardTitle>No Organizations Found</CardTitle>
                                <CardDescription>
                                    You are not assigned to any organizations
                                    yet. Please contact your organization
                                    administrator or create a new one.
                                </CardDescription>
                            </div>
                        </CardContent>
                    </Card>
                )}
            </div>
        </main>
    )
}
