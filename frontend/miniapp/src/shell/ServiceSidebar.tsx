// Application navigation composes the protected sidebar primitives and block parts.
"use client";

import * as React from "react";
import {
  ClipboardList,
  FolderOpen,
  Home,
  Layers,
  LayoutDashboard,
  List,
  MapPin,
  FileText,
  Copy,
  SquarePen,
  ClipboardPenLine,
} from "lucide-react";
import { RepositoryProvider, useRepositories } from "../business/repository/RepositoryContext";
import { repositoryNavigation } from "./repository-navigation";
import { NavMain, type NavItem } from "./sidebar-07/nav-main";
import { NavProjects } from "./sidebar-07/nav-projects";
import { NavUser, type AccountMenu } from "./sidebar-07/nav-user";
import { OrganizationSwitcher } from "@/shell/OrganizationSwitcher";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar,
  SidebarHeader,
  SidebarRail,
} from "@/shared/ui/sidebar";

type ServiceSidebarProps = React.ComponentProps<typeof Sidebar> & {
  workflow: string;
  account: AccountMenu;
  repositoryItem?: NavItem;
};

export function ServiceSidebar(props: ServiceSidebarProps) {
  // Projects remain independent of repository data; only their navigation owns a lazy catalog.
  if (props.workflow === "projects" && !props.repositoryItem) {
    return <RepositoryProvider workflow={props.workflow}><ProjectSidebar {...props} /></RepositoryProvider>;
  }
  return <NavigationSidebar {...props} />;
}

function ProjectSidebar(props: ServiceSidebarProps) {
  const spaces = useRepositories();
  return <NavigationSidebar {...props} repositoryItem={repositoryNavigation(spaces, props.workflow)} />;
}

function NavigationSidebar({ workflow, account, repositoryItem, ...props }: ServiceSidebarProps) {
  const { setOpenMobile } = useSidebar();
  const attendance = [
    { title: "出勤紀錄", url: "/miniapp/history", workflow: "history" },
    { title: "請假紀錄", url: "/miniapp/leave", workflow: "leave" },
    { title: "補卡與更正", url: "/miniapp/correction", workflow: "correction" },
    { title: "補卡審核", url: "/miniapp/review", workflow: "review" },
  ].map((item) => ({ ...item, isActive: item.workflow === workflow }));
  const items: NavItem[] = [
    {
      title: "首頁",
      url: "/miniapp",
      icon: Home,
      isActive: workflow === "home",
    },
    { title: "儀表板", url: "/miniapp/dashboard", icon: LayoutDashboard, isActive: workflow === "dashboard" },
    { title: "動態", url: "/miniapp/feed", icon: List, isActive: workflow === "feed" },
    { title: "專案", url: "/miniapp/projects", icon: Layers, isActive: workflow === "projects" },
    { ...(repositoryItem || {
      title: "儲存庫", url: "/miniapp/repositories", isActive: workflow === "repositories",
    }), icon: FolderOpen },
    {
      title: "出勤服務",
      icon: ClipboardList,
      isActive:
        attendance.some((item) => item.isActive) || workflow === "punch",
      items: attendance,
    },
  ];
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <OrganizationSwitcher />
      </SidebarHeader>
      <SidebarContent>
        <nav aria-label="快捷入口">
          <SidebarGroup>
            <SidebarGroupLabel>快捷入口</SidebarGroupLabel>
            <SidebarMenu>
              {[
                { title: "寫日誌", url: "/miniapp/feed?compose=1", icon: SquarePen },
                { title: "申請補卡", url: "/miniapp/correction", icon: ClipboardPenLine },
              ].map(({ title, url, icon: Icon }) => (
                <SidebarMenuItem key={url}>
                  <SidebarMenuButton asChild tooltip={title} onClick={() => setOpenMobile(false)}>
                    <a href={url} aria-label={title}><Icon aria-hidden="true" /><span>{title}</span></a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroup>
        </nav>
        <NavMain items={items} />
        <noscript><a href={repositoryItem?.url || "/miniapp/repositories"}>儲存庫（請啟用 JavaScript 載入清單）</a></noscript>
        <NavProjects
          projects={[
            { name: "物件範本", url: "/miniapp/object-templates", icon: Copy, isActive: workflow === "object-templates" },
            { name: "PDF 文字擷取", url: "/miniapp/poppler", icon: FileText, isActive: workflow === "poppler" },
            {
              name: "定位測試",
              url: "/location",
              icon: MapPin,
              isActive: workflow === "location",
            },
          ]}
        />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={account} />
      </SidebarFooter>
      <SidebarRail aria-label="切換側欄邊界" title="切換側欄" />
    </Sidebar>
  );
}
