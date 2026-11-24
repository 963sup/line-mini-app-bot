import React, { useState } from "react";
import Link from "next/link";
import Router from "next/router";
import { useOrganizationSelection } from "@/business/organization/OrganizationSelection";
import { Building2, ChevronsUpDown } from "lucide-react";
import { useOptionalSession } from "@/identity/Session";
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from "@/shared/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from "@/shared/ui/dropdown_menu";

export function OrganizationSwitcher() {
  const session = useOptionalSession();
  if (!session?.user) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            size="lg"
            className="group-data-[collapsible=icon]:justify-center"
            asChild
          >
            <div role="status" aria-label="組織切換狀態" aria-busy={session?.busy ?? false}>
              <Building2 aria-hidden="true" />
              <span className="group-data-[collapsible=icon]:hidden">
                {session?.busy ? "正在驗證身分…" : (session?.status ?? "尚未驗證身分")}
              </span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    );
  }
  return <Switcher key={session.epoch} />;
}

function Switcher() {
  const { isMobile, setOpenMobile } = useSidebar();
  const { context, error, saving, select } = useOrganizationSelection();
  const [failedHref, setFailedHref] = useState("");
  async function onSelected(organizationId: string) {
    const href = "/miniapp?" + new URLSearchParams({ organization: organizationId });
    setFailedHref("");
    setOpenMobile(false);
    try {
      if (!(await Router.push(href))) setFailedHref(href);
    } catch {
      setFailedHref(href);
    }
  }
  const label = (
    <>
      <Building2 />
      <span className="grid min-w-0 text-left group-data-[collapsible=icon]:hidden">
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
                className="group-data-[collapsible=icon]:justify-center"
                aria-label="切換組織"
                disabled={saving}
              >
                {label}
                <ChevronsUpDown
                  className="ml-auto group-data-[collapsible=icon]:hidden"
                  aria-hidden="true"
                />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent side={isMobile ? "bottom" : "right"} align="start">
              <DropdownMenuLabel>切換組織</DropdownMenuLabel>
              <DropdownMenuRadioGroup
                value={context.selected?.organizationId}
                onValueChange={(id) => select(id, onSelected)}
              >
                {context.groups.map((group) => (
                  <React.Fragment key={group.enterpriseId ?? "independent"}>
                    {group.enterpriseId ? (
                      <DropdownMenuItem asChild>
                        <Link
                          href={
                            "/miniapp?" +
                            new URLSearchParams({
                              enterprise: group.enterpriseId,
                            })
                          }
                        >
                          {group.label}
                        </Link>
                      </DropdownMenuItem>
                    ) : (
                      <DropdownMenuLabel>{group.label}</DropdownMenuLabel>
                    )}
                    {group.organizations.map((org) => (
                      <DropdownMenuRadioItem key={org.organizationId} value={org.organizationId}>
                        {org.name} {org.detail}
                      </DropdownMenuRadioItem>
                    ))}
                  </React.Fragment>
                ))}
              </DropdownMenuRadioGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        ) : context?.selected ? (
          <SidebarMenuButton
            size="lg"
            className="group-data-[collapsible=icon]:justify-center"
            asChild
          >
            <Link
              href={
                "/miniapp?" +
                new URLSearchParams({
                  organization: context.selected.organizationId,
                })
              }
            >
              {label}
            </Link>
          </SidebarMenuButton>
        ) : (
          <SidebarMenuButton
            size="lg"
            className="group-data-[collapsible=icon]:justify-center"
            asChild
          >
            <div role="status" aria-busy={!context && !error}>
              {label}
            </div>
          </SidebarMenuButton>
        )}
        {context?.enterprises?.map((ent) => (
          <SidebarMenuButton key={ent.id} className="group-data-[collapsible=icon]:hidden" asChild>
            <Link href={"/miniapp?" + new URLSearchParams({ enterprise: ent.id })}>
              <Building2 />
              {ent.name}
            </Link>
          </SidebarMenuButton>
        ))}
        {context?.canCreateEnterprise && (
          <SidebarMenuButton className="group-data-[collapsible=icon]:hidden" asChild>
            <Link href="/miniapp?enterprise=new">建立企業</Link>
          </SidebarMenuButton>
        )}
        {failedHref && (
          <Link className="group-data-[collapsible=icon]:hidden" href={failedHref}>
            重試開啟組織頁面
          </Link>
        )}
        {error && (
          <span className="group-data-[collapsible=icon]:hidden" role="alert">
            {error}
          </span>
        )}
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
