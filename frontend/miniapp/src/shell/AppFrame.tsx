import React from "react";
import Head from "next/head";
import Link from "next/link";
import type { NavItem } from "./sidebar-07/nav-main";
import { ServiceSidebar } from "./ServiceSidebar";
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
  account?: AccountMenu;
  repositoryItem?: NavItem;
  wide?: boolean;
};

export function AppFrame({
  title,
  workflow,
  children,
  account,
  repositoryItem,
  wide = false,
}: FrameProps) {
  const guest = {
    name: "定位診斷",
    status: "不需 LINE 登入",
    actions: [{ label: "返回正式服務", url: "/miniapp" }],
  };
  return (
    <SidebarProvider>
      <Head>
        <title>{`${title}｜LINE MINI App`}</title>
      </Head>
      <a href="#page-content" className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-50 focus:rounded-md focus:bg-background focus:p-3">跳至主要內容</a>
      <ServiceSidebar
        workflow={workflow}
        account={account || guest}
        repositoryItem={repositoryItem}
      />
      <SidebarInset className="min-w-0">
        <header className="flex min-h-14 items-center gap-3 border-b px-4 sm:px-6">
          <SidebarTrigger aria-label="切換側欄" />
          <Separator orientation="vertical" className="data-[orientation=vertical]:h-4" />
          <Breadcrumb aria-label="頁面導覽">
            <BreadcrumbList>
              {(workflow !== "home" || title !== "首頁") && <>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink asChild><Link href="/miniapp">首頁</Link></BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator className="hidden md:block" />
              </>}
              <BreadcrumbItem>
                <h1 className="text-base font-semibold"><BreadcrumbPage>{title}</BreadcrumbPage></h1>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div id="page-content" tabIndex={-1}
          className={cn(
            "mx-auto flex w-full min-w-0 flex-1 flex-col gap-4 px-4 py-4 pb-12 sm:px-6",
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
