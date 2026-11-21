import type { ReactNode } from "react";
import { OrganizationSelectionProvider } from "@/business/organization/OrganizationSelection";
import { SessionProvider } from "@/identity/Session";

export function MiniAppProviders({ children }: { children: ReactNode }) {
  return (
    <SessionProvider>
      <OrganizationSelectionProvider>{children}</OrganizationSelectionProvider>
    </SessionProvider>
  );
}

export function MiniAppLayout({ pathname, children }: { pathname: string; children: ReactNode }) {
  if (pathname === "/miniapp" || pathname.startsWith("/miniapp/"))
    return <MiniAppProviders>{children}</MiniAppProviders>;

  return children;
}
