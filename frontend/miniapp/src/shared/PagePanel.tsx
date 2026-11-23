import { cn } from "./utils";
import React from "react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "./ui/card";

export function PagePanel({
  heading,
  description,
  children,
  className,
  ...props
}: React.ComponentProps<typeof Card> & { heading: React.ReactNode; description?: React.ReactNode }) {
  return (
    <Card className={cn("min-w-0 gap-4 py-4 sm:gap-6 sm:py-6", className)} {...props}>
      <CardHeader className="min-w-0 px-4 sm:px-6">
        <CardTitle role="heading" aria-level={2} className="min-w-0 text-base leading-snug [overflow-wrap:anywhere]">
          {heading}
        </CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent className="flex min-w-0 flex-col gap-4 px-4 sm:px-6">{children}</CardContent>
    </Card>
  );
}
