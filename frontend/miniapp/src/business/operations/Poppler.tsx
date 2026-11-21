import React, { useRef, useState } from "react";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/shared/ui/field";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";

function encode(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result).split(",")[1]);
    reader.onerror = () => reject(new Error("無法讀取 PDF。"));
    reader.readAsDataURL(file);
  });
}

export function Poppler() {
  const { request, user } = useSession();
  const { run } = useLatestRequest();
  const locked = useRef(false);
  const [file, setFile] = useState<File>();
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [text, setText] = useState<string>();

  function extract(e: React.FormEvent) {
    e.preventDefault();
    if (!file || !user || locked.current) return;
    if (!file.size || file.size > 8 * 1024 * 1024) {
      setText(undefined); setNotice("請選擇 8 MiB 以內的非空 PDF。"); return;
    }
    locked.current = true; setBusy(true); setText(undefined); setNotice("");
    void run(async signal => {
      const data = await encode(file);
      if (signal.aborted) throw new DOMException("操作已取消", "AbortError");
      return request((client, s) => client.POST("/api/v1/tools/poppler/extractions", {
        signal: s, body: { filename: file.name, data },
      }), { signal, timeoutMs: 40000 });
    }, result => {
      locked.current = false; setBusy(false); setText(result.text);
      setNotice(result.text.trim() ? "文字擷取完成。" : "未擷取到文字；請核對原始 PDF，這不代表文件一定是掃描檔。");
    }, error => {
      locked.current = false; setBusy(false); setNotice(errorMessage(error));
    });
  }

  return <Card>
    <CardHeader>
      <CardTitle>PDF 文字擷取</CardTitle>
      <CardDescription>上傳 PDF 測試文字擷取與標準化。文件只用於本次處理，不儲存，也不送至 AI。</CardDescription>
    </CardHeader>
    <CardContent className="flex flex-col gap-4">
      <form onSubmit={extract}>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="poppler-file">PDF 文件</FieldLabel>
            <Input id="poppler-file" type="file" accept="application/pdf,.pdf" disabled={busy} onChange={e => {
              setFile(e.target.files?.[0]); setText(undefined); setNotice("");
            }} />
            <FieldDescription>每份文件最多 8 MiB，擷取時間上限 30 秒。</FieldDescription>
          </Field>
          <Field><Button type="submit" disabled={busy || !file || !user}>{busy ? "正在擷取……" : "擷取文字"}</Button></Field>
        </FieldGroup>
      </form>
      <p role="status" aria-live="polite">{notice}</p>
      {text !== undefined && <FieldGroup>
        <Field>
          <FieldLabel htmlFor="poppler-output">擷取結果</FieldLabel>
          <FieldDescription>保留空白與換行，分頁顯示為「⟦分頁⟧」。</FieldDescription>
          <Textarea id="poppler-output" readOnly value={text.replaceAll("\f", "\n⟦分頁⟧\n")} className="min-h-80 whitespace-pre font-mono" wrap="off" />
        </Field>
        <Field><Button type="button" variant="outline" onClick={() => {
          const url = URL.createObjectURL(new Blob([text], { type: "text/plain;charset=utf-8" }));
          const link = document.createElement("a"); link.href = url; link.download = "extracted.txt"; link.click();
          setTimeout(() => URL.revokeObjectURL(url), 1000);
        }}>下載文字原檔</Button></Field>
      </FieldGroup>}
    </CardContent>
  </Card>;
}
