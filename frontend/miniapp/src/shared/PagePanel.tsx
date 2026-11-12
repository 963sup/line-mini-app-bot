import React from "react";
import { Card, CardHeader, CardTitle, CardContent } from "./ui/card";
import { cn } from "./utils";

export function PagePanel({
  heading,
  children,
  className,
  ...props
}: React.ComponentProps<typeof Card> & { heading: React.ReactNode }) {
  return (
    <Card
      className={cn("mb-3 gap-4 rounded-2xl p-4 sm:p-6", className)}
      {...props}
    >
      <CardHeader className="px-0">
        <CardTitle role="heading" aria-level={2}>
          {heading}
        </CardTitle>
      </CardHeader>
      <CardContent className="px-0">{children}</CardContent>
    </Card>
  );
}
