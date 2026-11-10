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
      <h2>取得位置並完成打卡</h2>
      <p>請允許定位。取得位置後將保存本次打卡的位置、精度與採樣時間。</p>
      <button
        id="punch"
        disabled={punch.busy}
        aria-busy={punch.busy}
        onClick={() => void punch.start()}
      >
        取得位置並完成打卡
      </button>
      <p id="punch-status" role="status" aria-live="polite">
        {punch.status}
      </p>
    </section>
  );
}
