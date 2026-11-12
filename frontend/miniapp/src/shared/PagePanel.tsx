import React from "react";
import { Card, CardHeader, CardTitle, CardContent } from "./ui/card";

export function PagePanel({
  heading,
  children,
  className,
  ...props
}: React.ComponentProps<typeof Card> & { heading: React.ReactNode }) {
  return (
    <Card className={className} {...props}>
      <CardHeader>
        <CardTitle role="heading" aria-level={2}>
          {heading}
        </CardTitle>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}
