import React from "react";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import { useRepositories } from "./RepositoryContext";
import { displayTime } from "../../shared/time/time";

export function View() {
  const { selected, selectedID, scope, detailBusy, detailStatus, reloadDetail, canLoad } = useRepositories();
  return (
    <PagePanel heading="空間">
      {scope && <p className="text-sm text-muted-foreground">此連結僅開啟原群組空間，需具備該群組的存取資格。</p>}
      {selectedID ? <>
        <Button disabled={!canLoad || detailBusy} onClick={reloadDetail}>重新載入空間</Button>
        <Status role="status">{detailStatus}</Status>
      </> : <p>展開側邊欄「空間」，選擇可開啟的群組空間。</p>}
      <noscript>請啟用 JavaScript 並完成 LINE 身分驗證，以載入側邊欄空間及內容。</noscript>
      {selected && (
        <article>
          <h3 className="font-semibold">{selected.name || "群組空間"}</h3>
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
          <a
            className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring"
            href={
              "/miniapp/history?" +
              new URLSearchParams({ group: selected.groupId })
            }
          >
            查看本人出勤
          </a>
          <p className="text-sm leading-relaxed text-muted-foreground">
            摘要請回原群組操作。
          </p>
        </article>
      )}
    </PagePanel>
  );
}
