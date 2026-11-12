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
import { NavMain, type NavItem } from "./nav-main";
import { NavProjects } from "./nav-projects";
import { NavUser, type AccountMenu } from "./nav-user";
import { TeamSwitcher } from "./team-switcher";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/shared/ui/sidebar";

export function AppSidebar({
  workflow,
  contextItems,
  account,
  ...props
}: React.ComponentProps<typeof Sidebar> & {
  workflow: string;
  contextItems?: NavItem[];
  account: AccountMenu;
}) {
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
    {
      title: "空間",
      url: "/miniapp/repositories",
      icon: FolderOpen,
      isActive: workflow === "repositories",
    },
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
    ...(contextItems
      ? [
          {
            title: "示範空間",
            icon: Layers,
            isActive: true,
            items: contextItems,
          },
        ]
      : []),
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
              isActive: workflow !== "preview",
            },
            {
              name: "空間 UI 預覽",
              logo: Layers,
              plan: "示範資料",
              url: "/miniapp/workspace-preview",
              isActive: workflow === "preview",
            },
          ]}
        />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={items} />
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
