import React from "react";
import { Card, CardHeader, CardTitle, CardContent } from "./ui/card";

export function PagePanel({
  heading,
  children,
  ...props
}: React.ComponentProps<typeof Card> & { heading: React.ReactNode }) {
  return (
    <Card {...props}>
      <CardHeader className="min-w-0">
        <CardTitle role="heading" aria-level={2} className="min-w-0 break-words">
          {heading}
        </CardTitle>
      </CardHeader>
      <CardContent className="min-w-0">{children}</CardContent>
    </Card>
  );
}
