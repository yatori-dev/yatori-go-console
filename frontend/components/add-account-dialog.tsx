"use client"

import { AccountFormDialog } from "@/components/account-form-dialog"

type AddAccountDialogProps = {
    open: boolean
    onOpenChange: (open: boolean) => void
    onAdded: () => void
}

export function AddAccountDialog({ open, onOpenChange, onAdded }: AddAccountDialogProps) {
    return <AccountFormDialog open={open} onOpenChange={onOpenChange} onSaved={onAdded} />
}
