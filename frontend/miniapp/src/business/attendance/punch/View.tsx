import React from "react";
import { usePunch } from "./usePunch";
export function View() {
  const punch = usePunch();
  return (
    <section
      id="attendance"
      className="panel accent-panel"
      hidden={!punch.visible}
    >
      <h2>確認並完成打卡</h2>
      <p>先確認打卡狀態；需要新增紀錄時才取得位置，保存本次位置、精度與採樣時間。</p>
      <button
        id="punch"
        disabled={punch.busy}
        aria-busy={punch.busy}
        onClick={() => void punch.start()}
      >
        確認並完成打卡
      </button>
      <p id="punch-status" role="status" aria-live="polite">
        {punch.status}
      </p>
    </section>
  );
}
