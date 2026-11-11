import React from "react";
import Head from "next/head";
import type { ReactNode } from "react";
import { IdentityPanel } from "../identity/Panel";
import { SessionProvider, useSession } from "../identity/Session";
import { useEffect } from "react";
const navigationItems = [
  { workflow: "home", href: "/miniapp", label: "首頁", full: "首頁" },
  { workflow: "journal", href: "/miniapp/journal", label: "日誌", full: "日誌" },
  { workflow: "history", href: "/miniapp/history", label: "出勤", full: "本人月紀錄" },
  { workflow: "leave", href: "/miniapp/leave", label: "請假", full: "本人請假紀錄" },
  { workflow: "correction", href: "/miniapp/correction", label: "補登", full: "補卡／時間更正" },
  { workflow: "review", href: "/miniapp/review", label: "審核", full: "授權審核" },
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
  const currentPage = navigationItems.find((item) => item.workflow === workflow);
  const navigation = (
    <nav className={`page-nav${currentPage ? " compact-nav" : ""}`} aria-label="服務導覽">
      {navigationItems.map((item) => (
        <a key={item.workflow} href={item.href}
          aria-current={item.workflow === workflow ? "page" : undefined}
          title={item.full}
          aria-label={item.label !== item.full ? `${item.label}：${item.full}` : undefined}>
          {currentPage ? item.label : item.full}
        </a>
      ))}
    </nav>
  );
  const heading = <>
    <p className="eyebrow">{journal ? "LINE 日誌服務" : "LINE 出勤服務"}</p>
    <h1 title={title}>{currentPage?.label || title}</h1>
  </>;
  useEffect(() => {
    if (!ready || workflow !== "home") return;
    const params = new URLSearchParams(location.search);
    let leave = params.has("leave"),
      correction = params.get("correction"),
      journal = params.has("journal");
    try {
      const nested = params.get("liff.state");
      if (nested) {
        const query = new URL(nested, location.origin).searchParams;
        journal ||= query.has("journal");
        leave ||= query.has("leave");
        correction ||= query.get("correction");
      }
    } catch {}
    if (journal) location.replace("/miniapp/journal");
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
        <title>{title}｜{workflow === "journal" ? "LINE 日誌" : "LINE 出勤"}</title>
        <link rel="stylesheet" href="/assets/app.css" />
      </Head>
      <main className="app-shell">
        <header className={`app-hero${currentPage ? " compact-hero" : ""}`}>
          {currentPage ? <div className="compact-title">{heading}</div> : heading}
          {currentPage && navigation}
        </header>
        {!currentPage && workflow !== "punch" && navigation}
        {workflow !== "journal" && <IdentityPanel />}
        <div key={workflow === "journal" ? "journal" : epoch}>{children}</div>
        <footer className="app-footer">
          <p>
            {workflow === "journal"
              ? "日誌由 Google 表單處理，送出結果請以表單顯示為準。"
              : "時間均為臺灣時間。從群組打卡連結完成定位打卡；私訊 Bot 輸入「出勤紀錄」也可查詢。"}
          </p>
        </footer>
        <noscript>請啟用 JavaScript，以使用 LINE 服務。</noscript>
      </main>
    </>
  );
}
