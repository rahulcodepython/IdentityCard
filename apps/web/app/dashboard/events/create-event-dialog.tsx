"use client"

import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

import { CreateEventForm } from "./create-event-form"

export function CreateEventDialog({ allowFlash }: { allowFlash: boolean }) {
  const [open, setOpen] = useState(false)

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button>New event</Button>} />
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Create event</DialogTitle>
          <DialogDescription>
            Set up the schedule — you can add sub-events and people
            afterward.
          </DialogDescription>
        </DialogHeader>
        <CreateEventForm allowFlash={allowFlash} />
      </DialogContent>
    </Dialog>
  )
}
