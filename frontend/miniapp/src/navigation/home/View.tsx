import React from "react";
export function View() {
  return (
    <>
      <section className="panel">
        <h2>選擇要完成的事</h2>
        <div className="service-links">
          <a href="/miniapp/history">
            <strong>本人月紀錄</strong>
            <span>查看整月工時與班次明細，選擇班次提出更正。</span>
          </a>
          <a href="/miniapp/correction">
            <strong>補卡／時間更正</strong>
            <span>申請補卡或更正時間，查詢人工審核結果。</span>
          </a>
          <a href="/miniapp/leave">
            <strong>本人請假紀錄</strong>
            <span>查詢請假登記；新增申請請從群組輸入「請假」。</span>
          </a>
          <a href="/miniapp/review">
            <strong>授權審核</strong>
            <span>授權審核者查看群組申請，核准或填寫理由駁回。</span>
          </a>
        </div>
        <p className="hint">打卡請使用原群組的上班、下班或休息按鈕。</p>
      </section>
    </>
  );
}
