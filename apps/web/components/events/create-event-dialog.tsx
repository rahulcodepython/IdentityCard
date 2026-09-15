"use client";

import * as React from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { PlusIcon } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useCreateEventMutation } from "@/query-hooks/events.api";
import { CreateEventSchema, type CreateEventInput } from "@/schema/events.types";
import { toast } from "sonner";

interface CreateEventDialogProps {
    trigger?: React.ReactNode;
}

export function CreateEventDialog({ trigger }: CreateEventDialogProps) {
    const [open, setOpen] = React.useState(false);
    const createEventMutation = useCreateEventMutation();

    const form = useForm<CreateEventInput>({
        resolver: zodResolver(CreateEventSchema),
        defaultValues: {
            name: "",
            start_date: "",
            end_date: "",
        },
    });

    const handleOpenChange = (newOpen: boolean) => {
        setOpen(newOpen);
        if (!newOpen) {
            form.reset();
        }
    };

    const onSubmit = async (values: CreateEventInput) => {
        try {
            await createEventMutation.mutateAsync(values);
            setOpen(false);
            form.reset();
        } catch {
            toast.error("Failed to create event, Try again.");
        }
    };

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogTrigger render={
                trigger ? trigger as React.ReactElement
                    : <Button className="gap-1.5">
                        <PlusIcon className="size-4" />
                        <span>Create Event</span>
                    </Button>
            } />
            <DialogContent className="sm:max-w-md">
                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="flex flex-col gap-4">
                        <DialogHeader>
                            <DialogTitle>Create New Event</DialogTitle>
                            <DialogDescription>
                                Fill in the details below to create a new event.
                            </DialogDescription>
                        </DialogHeader>

                        <div className="flex flex-col gap-3">
                            <FormField
                                control={form.control}
                                name="name"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Event Name *</FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder="e.g. Annual Tech Summit"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <div className="grid grid-cols-2 gap-3">
                                <FormField
                                    control={form.control}
                                    name="start_date"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Start Date *</FormLabel>
                                            <FormControl>
                                                <Input type="date" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="end_date"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>End Date *</FormLabel>
                                            <FormControl>
                                                <Input type="date" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>
                        </div>

                        <DialogFooter className="mt-2">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setOpen(false)}
                                disabled={createEventMutation.isPending}
                            >
                                Cancel
                            </Button>
                            <Button type="submit" disabled={createEventMutation.isPending}>
                                {createEventMutation.isPending ? "Creating..." : "Create Event"}
                            </Button>
                        </DialogFooter>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
}
