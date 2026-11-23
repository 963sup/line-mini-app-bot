import { Alert, AlertDescription } from "./ui/alert";
import { Button } from "./ui/button";
import { Spinner } from "./ui/spinner";
import { Status } from "./Status";

/** Routine reads stay quiet; failures retain their existing accessible alert. */
export function ReadFeedback({
  busy,
  message,
  label = "正在讀取……",
  plain = false,
  onRetry,
}: {
  busy: boolean;
  message?: string;
  label?: string;
  plain?: boolean;
  onRetry?: () => void;
}) {
  if (busy)
    return (
      <p role="status" className="my-3 flex items-center gap-2 text-sm text-muted-foreground">
        <Spinner aria-hidden="true" />
        {label}
      </p>
    );
  if (plain && message && !onRetry)
    return (
      <p role="status" className="text-sm text-muted-foreground">
        {message}
      </p>
    );
  if (!onRetry) return message ? <Status>{message}</Status> : null;

  const retry = (
    <Button type="button" variant="outline" onClick={onRetry}>
      重試讀取
    </Button>
  );
  if (!message) return retry;
  return (
    <Alert variant="destructive" className="my-3">
      <AlertDescription>
        <p>{message}</p>
        {retry}
      </AlertDescription>
    </Alert>
  );
}
