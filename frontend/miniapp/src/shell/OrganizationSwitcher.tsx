import React from "react";
import { useOrganizationSelection } from "@/business/organization/useOrganizationSelection";
import { Building2, ChevronsUpDown } from "lucide-react";
import { useOptionalSession } from "@/identity/Session";
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
  DropdownMenuItem,
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
  const { isMobile, setOpenMobile } = useSidebar();
  const { context, error, saving, select } = useOrganizationSelection((organizationId) => {
    setOpenMobile(false);
    window.location.assign(
      "/miniapp?" + new URLSearchParams({ organization: organizationId }),
    );
  });
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
                    {group.enterpriseId ? (
                      <DropdownMenuItem asChild>
                        <a
                          href={
                            "/miniapp?" +
                            new URLSearchParams({
                              enterprise: group.enterpriseId,
                            })
                          }
                        >
                          {group.label}
                        </a>
                      </DropdownMenuItem>
                    ) : (
                      <DropdownMenuLabel>{group.label}</DropdownMenuLabel>
                    )}
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
        {context?.enterprises?.map((ent) => (
          <SidebarMenuButton key={ent.id} asChild>
            <a href={"/miniapp?" + new URLSearchParams({ enterprise: ent.id })}>
              <Building2 />
              {ent.name}
            </a>
          </SidebarMenuButton>
        ))}
        {context?.canCreateEnterprise && (
          <SidebarMenuButton asChild>
            <a href="/miniapp?enterprise=new">建立企業</a>
          </SidebarMenuButton>
        )}
        {error && <span role="alert">{error}</span>}
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
