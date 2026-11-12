import React from "react";
import { cn } from "./utils";
import { Alert, AlertDescription } from "./ui/alert";

/** Keep workflow status IDs and polite live announcements stable. */
export function Status({
  children,
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <Alert
      className={cn("my-4", className)}
      role="status"
      aria-live="polite"
      hidden={!children}
      {...props}
    >
      <AlertDescription>{children}</AlertDescription>
    </Alert>
  );
}
