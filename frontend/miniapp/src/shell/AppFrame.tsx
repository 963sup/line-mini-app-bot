import React from "react";
import Head from "next/head";
import Link from "next/link";
import type { NavItem } from "./sidebar-07/nav-main";
import { ServiceSidebar, type Shortcut } from "./ServiceSidebar";
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
import { SidebarProvider, SidebarInset, SidebarTrigger, useSidebar } from "@/shared/ui/sidebar";
import { cn } from "@/shared/utils";

type FrameProps = {
  title: string;
  workflow: string;
  children: React.ReactNode;
  account?: AccountMenu;
  repositoryItem?: NavItem;
  wide?: boolean;
  shortcut?: Shortcut;
};

function NavigationTrigger() {
  const { isMobile, openMobile, open } = useSidebar();
  const trigger = React.useRef<HTMLButtonElement>(null);
  const wasOpenMobile = React.useRef(false);
  React.useEffect(() => {
    if (isMobile && wasOpenMobile.current && !openMobile)
      trigger.current?.focus({ preventScroll: true });
    wasOpenMobile.current = openMobile;
  }, [isMobile, openMobile]);
  return (
    <SidebarTrigger
      ref={trigger}
      aria-label="切換側欄"
      aria-expanded={isMobile ? openMobile : open}
      className="size-11 shrink-0 md:size-9"
    />
  );
}

export function AppFrame({
  title,
  workflow,
  children,
  account,
  repositoryItem,
  wide = false,
  shortcut,
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
      <a
        href="#page-content"
        className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-50 focus:rounded-md focus:bg-background focus:p-3"
      >
        跳至主要內容
      </a>
      <ServiceSidebar
        workflow={workflow}
        shortcut={shortcut}
        account={account || guest}
        repositoryItem={repositoryItem}
      />
      <SidebarInset className="min-w-0 bg-muted">
        <header className="z-20 flex min-h-14 items-center gap-2 border-b bg-background px-3 max-md:sticky max-md:top-0 sm:gap-3 sm:px-6">
          <NavigationTrigger />
          <Separator orientation="vertical" className="data-[orientation=vertical]:h-4" />
          <Breadcrumb aria-label="頁面導覽">
            <BreadcrumbList>
              {(workflow !== "home" || title !== "首頁") && (
                <>
                  <BreadcrumbItem>
                    <BreadcrumbLink asChild className="inline-flex min-h-11 items-center">
                      <Link href="/miniapp">首頁</Link>
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                  <BreadcrumbSeparator />
                </>
              )}
              <BreadcrumbItem>
                <h1>
                  <BreadcrumbPage className="text-base font-semibold leading-snug">
                    {title}
                  </BreadcrumbPage>
                </h1>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div
          id="page-content"
          tabIndex={-1}
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
