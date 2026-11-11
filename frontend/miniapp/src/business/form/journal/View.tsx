import React from "react";

const formURL = "https://docs.google.com/forms/d/e/1FAIpQLSfK-_AVb1e-ZMgd9ITCQV_dwfI_d1nYheJauAUxx3vFvoXo9Q/viewform";

export function View() {
  return (
    <section className="panel journal-panel">
      <h2>填寫</h2>
      <p>內容由 Google 表單保存。</p>
      <p>
        無法顯示、需登入或上傳？
        <a href={formURL} target="_blank" rel="noopener noreferrer">開啟表單</a>。
      </p>
      <iframe className="journal-form" title="日誌 Google 表單" src={`${formURL}?embedded=true`} referrerPolicy="no-referrer" />
    </section>
  );
}
