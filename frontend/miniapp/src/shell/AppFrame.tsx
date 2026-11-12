import React from "react";
import Head from "next/head";
import type { NavItem } from "./sidebar-07/nav-main";
import type { LucideIcon } from "lucide-react";
import { AppSidebar } from "./sidebar-07/app-sidebar";
import type { AccountMenu } from "./sidebar-07/nav-user";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/shared/ui/breadcrumb";
import { Separator } from "@/shared/ui/separator";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/shared/ui/sidebar";
import { cn } from "@/shared/utils";

type FrameProps = {
  title: string;
  workflow: string;
  children: React.ReactNode;
  contextItems?: {
    label: string;
    icon: LucideIcon;
    onSelect?: () => void;
    active: boolean;
  }[];
  account?: AccountMenu;
  repositoryItem?: NavItem;
  wide?: boolean;
};

export function AppFrame({
  title,
  workflow,
  children,
  contextItems,
  account,
  repositoryItem,
  wide = false,
}: FrameProps) {
  const guest = {
    name: workflow === "preview" ? "預覽模式" : "定位診斷",
    status:
      workflow === "preview" ? "示範資料，不寫入正式系統" : "不需 LINE 登入",
    actions: [{ label: "返回正式服務", url: "/miniapp" }],
  };
  return (
    <SidebarProvider>
      <Head>
        <title>{title}｜LINE MINI App</title>
      </Head>
      <AppSidebar
        workflow={workflow}
        account={account || guest}
        repositoryItem={repositoryItem}
        contextItems={contextItems?.map((item) => ({
          title: item.label,
          icon: item.icon,
          isActive: item.active,
          onSelect: item.onSelect,
        }))}
      />
      <SidebarInset className="min-w-0">
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex min-w-0 items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" aria-label="切換側欄" />
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink href="/miniapp">LINE MINI App</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator className="hidden md:block" />
                <BreadcrumbItem>
                  <h1 className="text-sm font-medium">
                    <BreadcrumbPage>{title}</BreadcrumbPage>
                  </h1>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <div
          className={cn(
            "mx-auto flex w-full min-w-0 flex-1 flex-col gap-4 p-4 pt-0 pb-12",
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
