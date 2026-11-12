import React from "react";
import Head from "next/head";
import type { LucideIcon } from "lucide-react";
import {
  Building2,
  Home,
  FolderOpen,
  NotebookPen,
  ClipboardList,
  CalendarDays,
  FilePenLine,
  ClipboardCheck,
  MapPin,
  PanelsTopLeft,
} from "lucide-react";
import {
  SidebarProvider,
  Sidebar,
  SidebarHeader,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarInset,
  SidebarTrigger,
  useSidebar,
} from "@/shared/ui/sidebar";
import { cn } from "@/shared/utils";

type NavigationItem = {
  label: string;
  icon: LucideIcon;
  href?: string;
  onSelect?: () => void;
  active: boolean;
};
const services = [
  { workflow: "home", href: "/miniapp", label: "首頁", icon: Home },
  {
    workflow: "repositories",
    href: "/miniapp/repositories",
    label: "空間",
    icon: FolderOpen,
  },
  {
    workflow: "journal",
    href: "/miniapp/journal",
    label: "日誌",
    icon: NotebookPen,
  },
  {
    workflow: "history",
    href: "/miniapp/history",
    label: "出勤",
    icon: ClipboardList,
  },
  {
    workflow: "leave",
    href: "/miniapp/leave",
    label: "請假",
    icon: CalendarDays,
  },
  {
    workflow: "correction",
    href: "/miniapp/correction",
    label: "補登",
    icon: FilePenLine,
  },
  {
    workflow: "review",
    href: "/miniapp/review",
    label: "審核",
    icon: ClipboardCheck,
  },
];
const tools = [
  {
    workflow: "preview",
    href: "/miniapp/workspace-preview",
    label: "空間 UI 預覽",
    icon: PanelsTopLeft,
  },
  { workflow: "location", href: "/location", label: "定位測試", icon: MapPin },
];

type FrameProps = {
  title: string;
  workflow: string;
  children: React.ReactNode;
  contextItems?: NavigationItem[];
  wide?: boolean;
};

export function AppFrame({
  title,
  workflow,
  children,
  contextItems,
  wide = false,
}: FrameProps) {
  return (
    <SidebarProvider>
      <Head>
        <title>{title}｜LINE MINI App</title>
      </Head>
      <Sidebar collapsible="icon">
        <SidebarHeader>
          <div className="flex min-h-14 items-center gap-3 overflow-hidden px-2 text-primary">
            <Building2 className="size-6 shrink-0" aria-hidden="true" />
            <strong className="whitespace-nowrap group-data-[collapsible=icon]:hidden">
              LINE MINI App
            </strong>
          </div>
        </SidebarHeader>
        <SidebarContent>
          <NavigationGroup
            label="服務導覽"
            items={services.map((item) => ({
              ...item,
              active: workflow === item.workflow,
            }))}
          />
          {contextItems && (
            <NavigationGroup label="示範空間" items={contextItems} />
          )}
        </SidebarContent>
        <SidebarFooter>
          <NavigationGroup
            label="工具導覽"
            items={tools.map((item) => ({
              ...item,
              active: workflow === item.workflow,
            }))}
          />
        </SidebarFooter>
      </Sidebar>
      <SidebarInset className="min-w-0">
        <header className="flex min-h-16 items-center gap-3 border-b px-4 sm:px-6">
          <SidebarTrigger aria-label="切換側欄" />
          <h1 className="text-lg font-semibold tracking-tight">{title}</h1>
        </header>
        <div
          className={cn(
            "mx-auto flex w-full min-w-0 flex-col gap-3 px-3 py-5 pb-12 sm:px-6 sm:py-8",
            wide ? "max-w-[1360px]" : "max-w-[800px]",
          )}
        >
          {children}
          <noscript>請啟用 JavaScript，以使用 LINE 服務。</noscript>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

function NavigationGroup({
  label,
  items,
}: {
  label: string;
  items: NavigationItem[];
}) {
  const { setOpenMobile } = useSidebar();
  return (
    <nav aria-label={label}>
      <SidebarGroup>
        <SidebarGroupLabel>{label}</SidebarGroupLabel>
        <SidebarMenu>
          {items.map(({ label, icon: Icon, href, onSelect, active }) => (
            <SidebarMenuItem key={label}>
              <SidebarMenuButton
                asChild={!!href}
                isActive={active}
                tooltip={label}
                aria-current={active ? "page" : undefined}
                onClick={() => {
                  onSelect?.();
                  setOpenMobile(false);
                }}
              >
                {href ? (
                  <a href={href}>
                    <Icon aria-hidden="true" />
                    <span>{label}</span>
                  </a>
                ) : (
                  <>
                    <Icon aria-hidden="true" />
                    <span>{label}</span>
                  </>
                )}
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroup>
    </nav>
  );
}
