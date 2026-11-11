import React from "react";
import Head from "next/head";
import type { ReactNode } from "react";
import { IdentityPanel } from "../identity/Panel";
import { SessionProvider, useSession } from "../identity/Session";
import { useEffect } from "react";
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
  const navigation = (
    <nav className={`page-nav${journal ? " journal-nav" : ""}`} aria-label="服務導覽">
      <a href="/miniapp">首頁</a>
      <a href="/miniapp/journal" aria-current={journal ? "page" : undefined}>日誌</a>
      <a href="/miniapp/history" title={journal ? "本人月紀錄" : undefined} aria-label={journal ? "出勤：本人月紀錄" : undefined}>{journal ? "出勤" : "本人月紀錄"}</a>
      <a href="/miniapp/leave" title={journal ? "本人請假紀錄" : undefined} aria-label={journal ? "本人請假紀錄" : undefined}>{journal ? "請假" : "本人請假紀錄"}</a>
      <a href="/miniapp/correction" title={journal ? "補卡／時間更正" : undefined} aria-label={journal ? "補登：補卡／時間更正" : undefined}>{journal ? "補登" : "補卡／時間更正"}</a>
      <a href="/miniapp/review" title={journal ? "授權審核" : undefined} aria-label={journal ? "授權審核" : undefined}>{journal ? "審核" : "授權審核"}</a>
    </nav>
  );
  const heading = <>
    <p className="eyebrow">{journal ? "LINE 日誌服務" : "LINE 出勤服務"}</p>
    <h1>{title}</h1>
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
        <header className={`app-hero${journal ? " journal-hero" : ""}`}>
          {journal ? <div className="journal-title">{heading}</div> : heading}
          {journal && navigation}
        </header>
        {!journal && workflow !== "punch" && navigation}
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
