import React from "react";
import { Field, FieldLabel } from "./ui/field";
import { cn } from "./utils";

export function FormField({
  children,
  labelClassName,
  ...props
}: React.ComponentProps<typeof Field> & { labelClassName?: string }) {
  return (
    <Field {...props}>
      <FieldLabel
        className={cn(
          "w-full flex-col items-stretch gap-2 whitespace-normal",
          labelClassName,
        )}
      >
        {children}
      </FieldLabel>
    </Field>
  );
}
