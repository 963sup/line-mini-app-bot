import React, { useState } from "react";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { ToggleGroup, ToggleGroupItem } from "@/shared/ui/toggle-group";
import { useRepositories } from "./RepositoryContext";
import { navigation } from "./navigation";
import { displayTime } from "../../shared/time/time";
import { View as Issues } from "../issue/View";
import { View as Discussions } from "../discussion/View";
import { View as Milestones } from "./milestones/View";
import { View as Projects } from "../project/View";
import { View as Statistics } from "./statistics/View";
import { View as Labels } from "./labels/View";

export function View() {
  const { selected, selectedID, scope, detailBusy, detailStatus, reloadDetail, canLoad } = useRepositories();
  const [page, setPage] = useState("總覽");
  return (
    <section className="flex min-w-0 flex-col gap-4" aria-label="空間">
      {scope && <p className="text-sm text-muted-foreground">此連結僅開啟原群組空間，需具備該群組的存取資格。</p>}
      {selectedID ? <>
        <Button disabled={!canLoad || detailBusy} onClick={reloadDetail}>重新載入空間</Button>
        <Status role="status">{detailStatus}</Status>
      </> : <p>展開側邊欄「空間」，選擇可開啟的群組空間。</p>}
      <noscript>請啟用 JavaScript 並完成 LINE 身分驗證，以載入側邊欄空間及內容。</noscript>
      {selected && (
        <>
          <div>
            <h2 className="text-2xl font-semibold tracking-tight">{selected.name || "群組空間"}</h2>
            <p className="text-sm text-muted-foreground">在同一個空間掌握工作、交流想法與追蹤目標。</p>
          </div>
          <ToggleGroup type="single" value={page} onValueChange={(value) => { if (value) setPage(value); }} aria-label="空間功能" className="max-w-full flex-wrap justify-start">
            {navigation.map(({ title, icon: Icon }) => (
              <ToggleGroupItem key={title} value={title} aria-label={title}>
                <Icon aria-hidden="true" />{title}
              </ToggleGroupItem>
            ))}
          </ToggleGroup>
          {page === "總覽" ? (
            <div className="grid min-w-0 gap-4 lg:grid-cols-2">
              <Statistics key={selected.id} repositoryId={selected.id} />
              <PagePanel heading="空間資料">
                <dl className="grid gap-2 text-sm [&>dt]:font-medium [&>dd]:break-all [&>dd]:text-muted-foreground">
                  <dt>儲存庫 ID</dt>
                  <dd>{selected.id}</dd>
                  <dt>群組 ID</dt>
                  <dd>{selected.groupId}</dd>
                  <dt>狀態</dt>
                  <dd>{selected.status === "ACTIVE" ? "啟用" : "停用"}</dd>
                  <dt>版本</dt>
                  <dd>{selected.version}</dd>
                  <dt>建立時間</dt>
                  <dd>{displayTime(selected.createdAt)}</dd>
                  <dt>更新時間</dt>
                  <dd>{displayTime(selected.updatedAt)}</dd>
                </dl>
              </PagePanel>
              <PagePanel heading="出勤與群組服務">
                <a
                  className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring"
                  href={"/miniapp/history?" + new URLSearchParams({ group: selected.groupId })}
                >
                  查看本人出勤
                </a>
                <p className="text-sm leading-relaxed text-muted-foreground">摘要請回原群組操作。</p>
              </PagePanel>
            </div>
          ) : page === "事項" ? <Issues key={selected.id} repositoryId={selected.id} /> : page === "討論" ? <Discussions key={selected.id} repositoryId={selected.id} /> : page === "里程碑" ? <Milestones key={selected.id} repositoryId={selected.id} /> : page === "專案" ? <Projects key={selected.id} repositoryId={selected.id} /> : page === "標籤" ? <Labels key={selected.id} repositoryId={selected.id} /> : (
            <PagePanel heading={page}>
              <Empty>
                <EmptyHeader>
                  <EmptyTitle>{page}尚未開放</EmptyTitle>
                  <EmptyDescription>此空間目前提供基本資料與本人出勤查詢，協作功能將逐步開放。</EmptyDescription>
                </EmptyHeader>
              </Empty>
            </PagePanel>
          )}
        </>
      )}
    </section>
  );
}
