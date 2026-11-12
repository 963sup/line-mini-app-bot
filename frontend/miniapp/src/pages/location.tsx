import React from "react";
import Head from "next/head";
import { View } from "../diagnostics/geolocation/View";
export default function Page() {
  return (
    <>
      <Head>
        <title>取得目前位置</title>
      </Head>
      <main className="mx-auto my-8 flex max-w-2xl flex-col gap-5 px-4 py-8 sm:px-8">
        <p className="text-xs font-bold tracking-widest text-primary uppercase">
          WEB GEOLOCATION · Go + React
        </p>
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          取得目前位置
        </h1>
        <p>
          按下按鈕並允許定位後，會將經緯度、精度與採樣時間傳送至本站伺服器驗證。資料不會保存。
        </p>
        <View />
        <p className="text-sm leading-relaxed text-muted-foreground">
          定位可能受裝置、網路與權限影響；座標可被偽造，此頁不提供打卡或到場認證。
        </p>
        <p className="text-sm leading-relaxed text-muted-foreground">
          若拒絕權限，可在瀏覽器或系統設定允許定位後，再按一次按鈕。也可改用支援定位的外部瀏覽器。
        </p>
        <noscript>請啟用 JavaScript，以使用瀏覽器定位功能。</noscript>
      </main>
    </>
  );
}
