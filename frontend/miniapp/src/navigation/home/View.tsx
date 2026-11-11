import React from "react";
export function View() {
  return (
    <>
      <section className="panel">
        <h2>選擇要完成的事</h2>
        <div className="service-links">
          <a href="/miniapp/repositories">
            <strong>空間</strong>
            <span>查看群組儲存庫與本人出勤入口。</span>
          </a>
          <a href="/miniapp/journal">
            <strong>日誌</strong>
            <span>填寫 Google 表單。</span>
          </a>
          <a href="/miniapp/history">
            <strong>出勤</strong>
            <span>查月工時與班次，選擇班次更正。</span>
          </a>
          <a href="/miniapp/correction">
            <strong>補登</strong>
            <span>申請補卡、時間更正及查結果。</span>
          </a>
          <a href="/miniapp/leave">
            <strong>請假</strong>
            <span>查看登記；新增請在群組輸入「請假」。</span>
          </a>
          <a href="/miniapp/review">
            <strong>審核</strong>
            <span>授權人員核准或附理由駁回。</span>
          </a>
        </div>
        <p className="hint">上／下班及休息請用原群組按鈕。</p>
      </section>
    </>
  );
}
