import { Card, CardHeader, CardTitle, CardContent } from "@/shared/ui/card";
import React from "react";
export function View() {
  return (
    <>
      <Card className="mb-3 gap-4 rounded-2xl p-4 sm:p-6">
        <CardHeader className="px-0">
          <CardTitle role="heading" aria-level={2}>
            選擇要完成的事
          </CardTitle>
        </CardHeader>
        <CardContent className="px-0">
          <div className="my-3 grid grid-cols-2 gap-2 [&>a]:grid [&>a]:min-w-0 [&>a]:gap-1 [&>a]:rounded-xl [&>a]:border [&>a]:p-3 [&>a:hover]:bg-accent [&_span]:text-sm [&_span]:text-muted-foreground">
            <a className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring" href="/miniapp/repositories">
              <strong>空間</strong>
              <span>查看群組儲存庫與本人出勤入口。</span>
            </a>
            <a className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring" href="/miniapp/journal">
              <strong>日誌</strong>
              <span>填寫 Google 表單。</span>
            </a>
            <a className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring" href="/miniapp/history">
              <strong>出勤</strong>
              <span>查月工時與班次，選擇班次更正。</span>
            </a>
            <a className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring" href="/miniapp/correction">
              <strong>補登</strong>
              <span>申請補卡、時間更正及查結果。</span>
            </a>
            <a className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring" href="/miniapp/leave">
              <strong>請假</strong>
              <span>查看登記；新增請在群組輸入「請假」。</span>
            </a>
            <a className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring" href="/miniapp/review">
              <strong>審核</strong>
              <span>授權人員核准或附理由駁回。</span>
            </a>
          </div>
          <p className="text-sm leading-relaxed text-muted-foreground">上／下班及休息請用原群組按鈕。</p>
        </CardContent>
      </Card>
    </>
  );
}
