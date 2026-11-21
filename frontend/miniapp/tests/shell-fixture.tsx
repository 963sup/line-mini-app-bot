import type { ComponentProps } from "react";
import { Shell as MiniAppShell } from "../src/shell/Shell";
import { MiniAppProviders } from "../src/shell/MiniAppProviders";

export function Shell(props: ComponentProps<typeof MiniAppShell>) {
  return <MiniAppProviders><MiniAppShell {...props} /></MiniAppProviders>;
}
