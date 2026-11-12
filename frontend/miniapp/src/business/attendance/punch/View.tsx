import { Card, CardHeader, CardTitle, CardContent } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { Button } from "@/shared/ui/button";
import { Status } from "@/shared/Status";
import React from "react";
import { usePunch } from "./usePunch";
export function View() {
  const punch = usePunch();
  return (
    <Card
      id="attendance"
      className="mb-3 gap-4 rounded-2xl p-4 sm:p-6 bg-secondary"
      hidden={!punch.visible}
    >
      <CardHeader className="px-0">
        <CardTitle role="heading" aria-level={2}>
          確認並完成打卡
        </CardTitle>
      </CardHeader>
      <CardContent className="px-0">
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
      </CardContent>
    </Card>
  );
}
