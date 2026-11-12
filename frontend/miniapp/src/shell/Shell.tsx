import React from "react";
import Head from "next/head";
import type { ReactNode } from "react";
import { IdentityPanel } from "../identity/Panel";
import { SessionProvider, useSession } from "../identity/Session";
import { useEffect } from "react";
import { cn } from "@/shared/utils";
const navigationItems = [
  { workflow: "home", href: "/miniapp", label: "首頁", full: "首頁" },
  {
    workflow: "journal",
    href: "/miniapp/journal",
    label: "日誌",
    full: "日誌",
  },
  {
    workflow: "history",
    href: "/miniapp/history",
    label: "出勤",
    full: "本人月紀錄",
  },
  {
    workflow: "leave",
    href: "/miniapp/leave",
    label: "請假",
    full: "本人請假紀錄",
  },
  {
    workflow: "correction",
    href: "/miniapp/correction",
    label: "補登",
    full: "補卡／時間更正",
  },
  {
    workflow: "review",
    href: "/miniapp/review",
    label: "審核",
    full: "授權審核",
  },
];
export function Shell({
  title,
  workflow,
  children,
}: {
  title: string;
  workflow: string;
  children: ReactNode;
}) {
  return (
    <SessionProvider>
      <Content title={title} workflow={workflow}>
        {children}
      </Content>
    </SessionProvider>
  );
}
function Content({
  title,
  workflow,
  children,
}: {
  title: string;
  workflow: string;
  children: ReactNode;
}) {
  const { epoch, ready } = useSession();
  const journal = workflow === "journal";
  const currentPage =
    navigationItems.find((item) => item.workflow === workflow) ||
    (workflow === "repositories" ? { label: "空間" } : undefined);
  const navigation = (
    <nav
      className={cn(
        "grid min-w-0 flex-1 grid-cols-3 gap-1.5",
        currentPage ? "max-w-[360px]" : "mb-6",
      )}
      aria-label="服務導覽"
    >
      {navigationItems.map((item) => (
        <a
          className="flex min-h-11 items-center justify-center rounded-lg bg-secondary px-2 py-1 text-sm text-secondary-foreground no-underline hover:bg-accent aria-[current=page]:bg-accent aria-[current=page]:font-bold"
          key={item.workflow}
          href={item.href}
          aria-current={item.workflow === workflow ? "page" : undefined}
          title={item.full}
          aria-label={
            item.label !== item.full ? `${item.label}：${item.full}` : undefined
          }
        >
          {currentPage ? item.label : item.full}
        </a>
      ))}
    </nav>
  );
  const heading = (
    <>
      <p className="text-xs font-bold tracking-widest text-primary uppercase">
        {journal ? "LINE 日誌服務" : "LINE 出勤服務"}
      </p>
      <h1
        className="my-1 text-3xl font-bold tracking-tight sm:text-4xl"
        title={title}
      >
        {currentPage?.label || title}
      </h1>
    </>
  );
  useEffect(() => {
    if (!ready || workflow !== "home") return;
    const params = new URLSearchParams(location.search);
    let leave = params.has("leave"),
      correction = params.get("correction"),
      journal = params.has("journal"),
      repositories = params.has("repositories"),
      repository = params.get("repository") || "";
    try {
      const nested = params.get("liff.state");
      if (nested) {
        const query = new URL(nested, location.origin).searchParams;
        repositories ||= query.has("repositories");
        repository ||= query.get("repository") || "";
        journal ||= query.has("journal");
        leave ||= query.has("leave");
        correction ||= query.get("correction");
      }
    } catch {}
    if (repositories)
      location.replace(
        "/miniapp/repositories" +
          (repository ? "?" + new URLSearchParams({ id: repository }) : ""),
      );
    else if (journal) location.replace("/miniapp/journal");
    else if (correction)
      location.replace(
        "/miniapp/correction?" + new URLSearchParams({ group: correction }),
      );
    else if (leave) location.replace("/miniapp/leave" + location.search);
    else if (params.has("launch") || params.has("liff.state"))
      location.replace("/miniapp/punch" + location.search);
  }, [workflow, ready]);
  return (
    <>
      <Head>
        <title>
          {title}｜{workflow === "journal" ? "LINE 日誌" : "LINE 出勤"}
        </title>
      </Head>
      <main className="mx-auto max-w-[760px] px-3 py-4 pb-12 sm:px-5 sm:py-10 sm:pb-18">
        <header
          className={cn(
            "mb-4 flex gap-4 px-1.5 py-2",
            currentPage ? "items-center justify-between" : "flex-col",
          )}
        >
          {currentPage ? <div className="shrink-0">{heading}</div> : heading}
          {currentPage && navigation}
        </header>
        {!currentPage && workflow !== "punch" && navigation}
        {workflow !== "journal" && <IdentityPanel />}
        <div
          className="flex flex-col gap-3"
          key={workflow === "journal" ? "journal" : epoch}
        >
          {children}
        </div>
        <noscript>請啟用 JavaScript，以使用 LINE 服務。</noscript>
      </main>
    </>
  );
}
