import React, { useEffect, useState } from "react";
import type { components } from "@repo/api-client";
import { Building2, ChevronsUpDown } from "lucide-react";
import { useOptionalSession, useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/shared/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from "@/shared/ui/dropdown-menu";

import { TeamSwitcher } from "./sidebar-07/team-switcher";

export function OrganizationSwitcher() {
  const session = useOptionalSession();
  if (!session?.user)
    return (
      <TeamSwitcher
        teams={[
          {
            name: "LINE MINI App",
            logo: Building2,
            plan: "正式服務",
            url: "/miniapp",
            isActive: true,
          },
        ]}
      />
    );
  return <Switcher key={session.epoch} />;
}

function Switcher() {
  const { request } = useSession(),
    { isMobile, setOpenMobile } = useSidebar(),
    { run, cancel } = useLatestRequest();
  const [context, setContext] =
    useState<components["schemas"]["OrganizationContext"]>();
  const [error, setError] = useState(""),
    [saving, setSaving] = useState(false);
  function load() {
    return run(
      (signal) =>
        request(
          (api, s) => api.GET("/api/v1/me/organization-context", { signal: s }),
          { signal },
        ),
      setContext,
      (e) => setError(errorMessage(e)),
    );
  }
  useEffect(() => {
    void load();
    return cancel;
  }, [request, run, cancel]);
  function select(organizationId: string) {
    if (saving || context?.selected?.organizationId === organizationId) return;
    setSaving(true);
    setError("");
    void run(
      (signal) =>
        request(
          (api, s) =>
            api.PUT("/api/v1/me/organization-context", {
              signal: s,
              body: { organizationId },
            }),
          { signal },
        ),
      (next) => {
        setContext(next);
        setOpenMobile(false);
        window.location.assign(
          "/miniapp?" + new URLSearchParams({ organization: organizationId }),
        );
      },
      (e) => {
        setSaving(false);
        setError(errorMessage(e));
        void load();
      },
    );
  }
  const label = (
    <>
      <Building2 />
      <span className="grid min-w-0 text-left">
        <span className="truncate">
          {context?.selected?.name ??
            (context ? "尚無可用組織" : error ? "無法載入組織" : "載入組織中…")}
        </span>
        <small className="truncate">{context?.selected?.groupLabel}</small>
      </span>
    </>
  );
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        {context?.canSwitch ? (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <SidebarMenuButton
                size="lg"
                aria-label="切換組織"
                disabled={saving}
              >
                {label}
                <ChevronsUpDown />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              side={isMobile ? "bottom" : "right"}
              align="start"
            >
              <DropdownMenuLabel>切換組織</DropdownMenuLabel>
              <DropdownMenuRadioGroup
                value={context.selected?.organizationId}
                onValueChange={select}
              >
                {context.groups.map((group) => (
                  <React.Fragment key={group.enterpriseId ?? "independent"}>
                    <DropdownMenuLabel>{group.label}</DropdownMenuLabel>
                    {group.organizations.map((org) => (
                      <DropdownMenuRadioItem
                        key={org.organizationId}
                        value={org.organizationId}
                      >
                        {org.name} {org.detail}
                      </DropdownMenuRadioItem>
                    ))}
                  </React.Fragment>
                ))}
              </DropdownMenuRadioGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        ) : (
          <SidebarMenuButton size="lg" asChild>
            <a
              href={
                context?.selected
                  ? "/miniapp?" +
                    new URLSearchParams({
                      organization: context.selected.organizationId,
                    })
                  : undefined
              }
            >
              {label}
            </a>
          </SidebarMenuButton>
        )}
        {error && <span role="alert">{error}</span>}
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
