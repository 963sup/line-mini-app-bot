import React, { useRef, useState } from "react";
import { Button } from "./ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "./ui/dialog";

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
  const dialog = <Dialog open={open} onOpenChange={value => { if (!value) keepEditing(); }}>
    <DialogContent showCloseButton={false} onCloseAutoFocus={event => {
      event.preventDefault();
      if (!approved.current) focusEditor();
    }}>
      <DialogHeader>
        <DialogTitle>放棄未儲存的修改？</DialogTitle>
        <DialogDescription>修改尚未儲存。放棄後將保留原有資料。</DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button type="button" variant="outline" onClick={keepEditing}>繼續編輯</Button>
        <Button type="button" variant="destructive" onClick={() => {
          const next = action.current;
          approved.current = true; action.current = null; setOpen(false); next?.();
        }}>放棄修改</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>;
  return { navigate, dialog, isConfirming: () => action.current !== null };
}
