import { PagePanel } from "@/shared/PagePanel";
import { Spinner } from "@/shared/ui/spinner";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import React from "react";
import { usePunch } from "./usePunch";
export function View() {
  const punch = usePunch();
  return (
    <PagePanel
      id="attendance"
      className="bg-secondary"
      hidden={!punch.visible}
      heading="確認並完成打卡"
    >
      <p>
        先確認打卡狀態；需要新增紀錄時才取得位置，保存本次位置、精度與採樣時間。
      </p>
      <Button
        id="punch"
        disabled={punch.busy}
        aria-busy={punch.busy}
        onClick={() => void punch.start()}
      >
        {punch.busy && <Spinner aria-label="處理中" />}
        確認並完成打卡
      </Button>
      <Status id="punch-status" role="status" aria-live="polite">
        {punch.status}
      </Status>
    </PagePanel>
  );
}
