import React from "react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "./ui/card";

export function PagePanel({
  heading,
  description,
  children,
  ...props
}: React.ComponentProps<typeof Card> & { heading: React.ReactNode; description?: React.ReactNode }) {
  return (
    <Card {...props}>
      <CardHeader className="min-w-0">
        <CardTitle role="heading" aria-level={2} className="min-w-0 break-words">
          {heading}
        </CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent className="flex min-w-0 flex-col gap-4">{children}</CardContent>
    </Card>
  );
}
