// Application navigation composes the protected sidebar primitives and block parts.
"use client";

import * as React from "react";
import {
  type LucideIcon,
  ClipboardList,
  FolderOpen,
  Home,
  Layers,
  LayoutDashboard,
  List,
  MapPin,
  FileText,
  Copy,
  X,
  SquarePen,
  ClipboardPenLine,
} from "lucide-react";
import { RepositoryProvider, useRepositories } from "../business/repository/RepositoryContext";
import { repositoryNavigation } from "./repository_navigation";
import { NavMain, type NavItem } from "./navigation/nav_main";
import { Button } from "@/shared/ui/button";
import { NavProjects } from "./navigation/nav_projects";
import { NavUser, type AccountMenu } from "./navigation/nav_user";
import { OrganizationSwitcher } from "@/shell/OrganizationSwitcher";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
  SidebarHeader,
  SidebarRail,
} from "@/shared/ui/sidebar";

export type Shortcut = "compose" | "correction";

type Destination = { name: string; url: string; icon: LucideIcon; isActive?: boolean };

function DestinationGroup({
  label,
  navLabel,
  links,
}: {
  label: string;
  navLabel: string;
  links: Destination[];
}) {
  const { setOpenMobile } = useSidebar();
  return (
    <nav aria-label={navLabel}>
      <SidebarGroup>
        <SidebarGroupLabel>{label}</SidebarGroupLabel>
        <SidebarMenu>
          {links.map(({ name, url, icon: Icon, isActive }) => (
            <SidebarMenuItem key={url}>
              <SidebarMenuButton
                asChild
                tooltip={name}
                isActive={isActive}
                onClick={() => setOpenMobile(false)}
              >
                <a href={url} aria-label={name} aria-current={isActive ? "page" : undefined}>
                  <Icon aria-hidden="true" />
                  <span>{name}</span>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroup>
    </nav>
  );
}

type ServiceSidebarProps = React.ComponentProps<typeof Sidebar> & {
  workflow: string;
  shortcut?: Shortcut;
  account: AccountMenu;
  repositoryItem?: NavItem;
};

export function ServiceSidebar(props: ServiceSidebarProps) {
  // Projects remain independent of repository data; only their navigation owns a lazy catalog.
  if (props.workflow === "projects" && !props.repositoryItem) {
    return (
      <RepositoryProvider workflow={props.workflow}>
        <ProjectSidebar {...props} />
      </RepositoryProvider>
    );
  }
  return <NavigationSidebar {...props} />;
}

function ProjectSidebar(props: ServiceSidebarProps) {
  const spaces = useRepositories();
  return (
    <NavigationSidebar {...props} repositoryItem={repositoryNavigation(spaces, props.workflow)} />
  );
}

function NavigationSidebar({
  workflow,
  shortcut,
  account,
  repositoryItem,
  ...props
}: ServiceSidebarProps) {
  const { state, isMobile, setOpenMobile } = useSidebar();
  const items: NavItem[] = [
    {
      title: "首頁",
      url: "/miniapp",
      icon: Home,
      isActive: workflow === "home",
    },
    {
      title: "儀表板",
      url: "/miniapp/dashboard",
      icon: LayoutDashboard,
      isActive: workflow === "dashboard",
    },
    { title: "動態", url: "/miniapp/feed", icon: List, isActive: workflow === "feed" && !shortcut },
    { title: "專案", url: "/miniapp/projects", icon: Layers, isActive: workflow === "projects" },
    {
      ...(repositoryItem || {
        title: "儲存庫",
        url: "/miniapp/repositories",
        isActive: workflow === "repositories",
      }),
      icon: FolderOpen,
    },
    {
      title: "出勤服務",
      icon: ClipboardList,
      url: "/miniapp/attendance",
      isActive: (workflow === "attendance" || workflow === "punch") && !shortcut,
    },
  ];
  const tools = [
    {
      name: "從文件辨識物件",
      url: "/miniapp/object-recognition",
      icon: FileText,
      isActive: workflow === "object-recognition",
    },
    {
      name: "物件範本",
      url: "/miniapp/object-templates",
      icon: Copy,
      isActive: workflow === "object-templates",
    },
    {
      name: "PDF 文字擷取",
      url: "/miniapp/poppler",
      icon: FileText,
      isActive: workflow === "poppler",
    },
    {
      name: "定位測試",
      url: "/location",
      icon: MapPin,
      isActive: workflow === "location",
    },
  ];
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        {isMobile && (
          <div className="flex items-center justify-between px-2">
            <span className="font-semibold">服務導覽</span>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-11"
              aria-label="關閉側欄"
              onClick={() => setOpenMobile(false)}
            >
              <X aria-hidden="true" />
            </Button>
          </div>
        )}
        <OrganizationSwitcher />
      </SidebarHeader>
      <SidebarContent className="group-data-[collapsible=icon]:overflow-y-auto group-data-[collapsible=icon]:overflow-x-hidden">
        <DestinationGroup
          label="快捷入口"
          navLabel="快捷入口"
          links={[
            {
              name: "寫日誌",
              url: "/miniapp/feed?compose=1",
              icon: SquarePen,
              isActive: shortcut === "compose",
            },
            {
              name: "申請補卡",
              url: "/miniapp/attendance?tab=correction",
              icon: ClipboardPenLine,
              isActive: shortcut === "correction",
            },
          ]}
        />
        <NavMain items={items} />
        <noscript>
          <a href={repositoryItem?.url || "/miniapp/repositories"}>
            儲存庫（請啟用 JavaScript 載入清單）
          </a>
        </noscript>
        {state === "collapsed" && !isMobile ? (
          <DestinationGroup label="工具" navLabel="工具導覽" links={tools} />
        ) : (
          <NavProjects projects={tools} />
        )}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={account} />
      </SidebarFooter>
      <SidebarRail aria-label="切換側欄邊界" title="切換側欄" />
    </Sidebar>
  );
}
