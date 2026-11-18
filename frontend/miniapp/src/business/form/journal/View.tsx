import { Card } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import React from "react";

const formURL =
  "https://docs.google.com/forms/d/e/1FAIpQLSfK-_AVb1e-ZMgd9ITCQV_dwfI_d1nYheJauAUxx3vFvoXo9Q/viewform";

export function View() {
  return (
    <Card className="mb-3 gap-4 rounded-2xl p-4 sm:p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-semibold">填寫</h2>
        <Button variant="outline" asChild>
          <a className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring" href={formURL} target="_blank" rel="noopener noreferrer">
            開啟表單
          </a>
        </Button>
      </div>
      <iframe
        className="mt-4 block h-[80vh] min-h-[600px] w-full border-0"
        title="日誌 Google 表單"
        src={`${formURL}?embedded=true`}
        referrerPolicy="no-referrer"
      />
    </Card>
  );
}
