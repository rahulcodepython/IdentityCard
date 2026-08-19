"use client"

import { useState } from "react"

// The create/edit/delete dialog-state triplet (template §1.2/1.6),
// extracted once it's copy-pasted across pages — every CRUD screen wires
// its DataTable + FormDialog + ConfirmDeleteDialog off one of these.
export function useCrudDialogState<T>() {
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<T | null>(null)
  const [deleting, setDeleting] = useState<T | null>(null)

  return {
    createOpen,
    openCreate: () => setCreateOpen(true),
    closeCreate: () => setCreateOpen(false),

    editing,
    openEdit: (item: T) => setEditing(item),
    closeEdit: () => setEditing(null),

    deleting,
    requestDelete: (item: T) => setDeleting(item),
    cancelDelete: () => setDeleting(null),
    confirmDelete: async (onDelete: (item: T) => Promise<void> | void) => {
      if (!deleting) return
      await onDelete(deleting)
      setDeleting(null)
    },
  }
}
