import React from "react";
import { Button } from "./ui/button";
import { Spinner } from "./ui/spinner";
import { Status } from "./Status";

/** Routine reads stay quiet; failures retain their existing accessible alert. */
export function ReadFeedback({ busy, message, label = "正在讀取……", onRetry }: { busy: boolean; message?: string; label?: string; onRetry?: () => void }) {
  return busy ? <p role="status" className="my-3 flex items-center gap-2 text-sm text-muted-foreground"><Spinner aria-hidden="true" />{label}</p> : <><Status>{message}</Status>{onRetry && <Button type="button" variant="outline" onClick={onRetry}>重試讀取</Button>}</>;
}
