import React, { useRef, useState } from "react";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "./ui/alert-dialog";

/** Keep a single pending navigation; never submit or replace it while confirming. */
export function useDiscardChanges(dirty: boolean, focusEditor: () => void) {
  const action = useRef<(() => void) | null>(null);
  const approved = useRef(false);
  const [open, setOpen] = useState(false);
  function navigate(next: () => void) {
    if (action.current) return;
    if (!dirty) { next(); return; }
    action.current = next; approved.current = false; setOpen(true);
  }
  function keepEditing() {
    action.current = null; setOpen(false);
  }
  const dialog = <AlertDialog open={open} onOpenChange={value => { if (!value) keepEditing(); }}>
    <AlertDialogContent onCloseAutoFocus={event => {
      event.preventDefault();
      if (!approved.current) focusEditor();
    }}>
      <AlertDialogHeader>
        <AlertDialogTitle>放棄未儲存的修改？</AlertDialogTitle>
        <AlertDialogDescription>修改尚未儲存。放棄後將保留原有資料。</AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>繼續編輯</AlertDialogCancel>
        <AlertDialogAction variant="destructive" onClick={() => {
          const next = action.current;
          approved.current = true; action.current = null; setOpen(false); next?.();
        }}>放棄修改</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>;
  return { navigate, dialog, isConfirming: () => action.current !== null };
}
