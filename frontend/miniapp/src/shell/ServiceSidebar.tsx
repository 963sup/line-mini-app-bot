// Application navigation composes the protected sidebar primitives and block parts.
"use client";

import * as React from "react";
import {
  Building2,
  ClipboardList,
  FolderOpen,
  Home,
  Layers,
  MapPin,
  NotebookPen,
} from "lucide-react";
import { RepositoryProvider, useRepositories } from "../business/repository/RepositoryContext";
import { repositoryNavigation } from "./repository-navigation";
import { NavMain, type NavItem } from "./sidebar-07/nav-main";
import { NavProjects } from "./sidebar-07/nav-projects";
import { NavUser, type AccountMenu } from "./sidebar-07/nav-user";
import { TeamSwitcher } from "./sidebar-07/team-switcher";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
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
  const attendance = [
    { title: "出勤", url: "/miniapp/history", workflow: "history" },
    { title: "請假", url: "/miniapp/leave", workflow: "leave" },
    { title: "補登", url: "/miniapp/correction", workflow: "correction" },
    { title: "審核", url: "/miniapp/review", workflow: "review" },
  ].map((item) => ({ ...item, isActive: item.workflow === workflow }));
  const items: NavItem[] = [
    {
      title: "首頁",
      url: "/miniapp",
      icon: Home,
      isActive: workflow === "home",
    },
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
    {
      title: "日誌",
      url: "/miniapp/journal",
      icon: NotebookPen,
      isActive: workflow === "journal",
    },
  ];
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
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
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={items} />
        <noscript><a href={repositoryItem?.url || "/miniapp/repositories"}>儲存庫（請啟用 JavaScript 載入清單）</a></noscript>
        <NavProjects
          projects={[
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
