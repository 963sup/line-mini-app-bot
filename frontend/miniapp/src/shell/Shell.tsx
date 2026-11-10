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
  useEffect(() => {
    if (
      ready &&
      workflow === "home" &&
      (new URLSearchParams(location.search).has("launch") ||
        new URLSearchParams(location.search).has("liff.state"))
    )
      location.replace("/miniapp/punch" + location.search);
  }, [workflow, ready]);
  return (
    <>
      <Head>
        <title>{title}｜LINE 出勤</title>
        <link rel="stylesheet" href="/assets/app.css" />
      </Head>
      <main className="app-shell">
        <header className="app-hero">
          <p className="eyebrow">LINE 出勤服務</p>
          <h1>{title}</h1>
        </header>
        {workflow !== "punch" && (
          <nav className="page-nav" aria-label="出勤服務">
            <a href="/miniapp">首頁</a>
            <a href="/miniapp/history">本人月紀錄</a>
            <a href="/miniapp/correction">更正與申請結果</a>
            <a href="/miniapp/review">授權審核</a>
          </nav>
        )}
        <IdentityPanel />
        <div key={epoch}>{children}</div>
        <footer className="app-footer">
          <p>
            時間均為臺灣時間。從群組打卡連結完成定位打卡；私訊 Bot
            輸入「出勤紀錄」也可查詢。
          </p>
        </footer>
        <noscript>請啟用 JavaScript，以使用 LINE 登入及出勤服務。</noscript>
      </main>
    </>
  );
}
