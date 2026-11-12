import React from "react";
import { PagePanel } from "@/shared/PagePanel";

export function View() {
  return (
    <PagePanel heading="開始使用">
      <p>從側欄選擇空間、日誌或出勤功能；手機可點左上角開啟導覽。</p>
      <p className="text-sm text-muted-foreground">
        上／下班及休息請用原群組按鈕。
      </p>
    </PagePanel>
  );
}
