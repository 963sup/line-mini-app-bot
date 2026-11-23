import { useRef, useState } from "react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { FormField } from "@/shared/FormField";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";

import { FieldGroup } from "@/shared/ui/field";

import { encodeFile } from "./Candidates";

import {
  Trace,
  TraceCollapsible,
  TraceCollapsibleContent,
  TraceCollapsibleTrigger,
  TraceLine,
  TraceTitle,
} from "@/shared/ui/trace";
type Result = components["schemas"]["ObjectTemplateDraftResult"];

export type Review = { result: Result; source: "document" | "procedure" };

export function Evidence({
  candidate,
  source,
}: {
  candidate: Result["candidates"][number];
  source: Review["source"];
}) {
  return (
    <Trace>
      <TraceCollapsible defaultOpen>
        <TraceCollapsibleTrigger type="button">
          <TraceLine>
            <TraceTitle>{source === "document" ? "文件依據" : "工序依據"}</TraceTitle>
          </TraceLine>
        </TraceCollapsibleTrigger>
        <TraceCollapsibleContent>
          {source === "document" && (
            <p>
              {candidate.quantity ?? "數量待核對"} {candidate.unit ?? "單位待核對"} ·{" "}
              {candidate.location ?? "位置待核對"}
            </p>
          )}
          <p>
            {source === "document" ? "原文：" : "推導說明："}
            {candidate.quote}
          </p>
        </TraceCollapsibleContent>
      </TraceCollapsible>
    </Trace>
  );
}

export function TemplatePreparation({
  source,
  disabled,
  onDraft,
}: {
  source: Review["source"];
  disabled: boolean;
  onDraft: (review: Review) => void;
}) {
  const { request } = useSession();
  const { run, cancel } = useLatestRequest();
  const [file, setFile] = useState<File>();
  const [name, setName] = useState("");
  const [conditions, setConditions] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const locked = useRef(false);
  function prepare() {
    if (disabled || locked.current) return;
    if (source === "document" && (!file || file.size === 0 || file.size > 8 * 1024 * 1024)) {
      setMessage("請選擇 8 MiB 以內的非空文件。");
      return;
    }
    if (source === "procedure" && !name.trim()) {
      setMessage("請填寫工項名稱。");
      return;
    }
    locked.current = true;
    setBusy(true);
    setMessage("");
    void run(
      async (signal) => {
        const body =
          source === "document"
            ? { source, filename: file!.name, data: await encodeFile(file!) }
            : { source, name, conditions };
        if (signal.aborted) throw new DOMException("操作已取消", "AbortError");
        return request(
          (api, s) => api.POST("/api/v1/object-template-drafts", { signal: s, body }),
          { signal, timeoutMs: 95000 },
        );
      },
      (result) => {
        locked.current = false;
        setBusy(false);
        onDraft({ source, result });
      },
      (error) => {
        locked.current = false;
        setBusy(false);
        setMessage(errorMessage(error));
      },
    );
  }
  return (
    <FieldGroup>
      <p className="text-sm text-muted-foreground">
        按下產生時，文件或工項條件會傳送至 OpenAI。結果只供覆核，確認建立範本後才儲存。
      </p>
      {source === "document" ? (
        <FormField>
          來源文件
          <Input
            type="file"
            disabled={busy || disabled}
            onChange={(e) => {
              setFile(e.target.files?.[0]);
            }}
          />
        </FormField>
      ) : (
        <>
          <FormField>
            工項名稱
            <Input
              value={name}
              maxLength={120}
              disabled={busy || disabled}
              onChange={(e) => {
                setName(e.target.value);
              }}
            />
          </FormField>
          <FormField>
            施工條件
            <Textarea
              value={conditions}
              maxLength={4000}
              disabled={busy || disabled}
              placeholder="例如工法、材料、施工環境與限制"
              onChange={(e) => {
                setConditions(e.target.value);
              }}
            />
          </FormField>
        </>
      )}
      <Button type="button" disabled={busy || disabled} onClick={prepare}>
        {source === "document" ? "辨識候選" : "推導工序"}
      </Button>
      {busy && (
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            cancel();
            locked.current = false;
            setBusy(false);
            setMessage("已取消等待，可重新產生。");
          }}
        >
          取消等待
        </Button>
      )}
      <p role="status">{busy ? "正在產生草稿……" : message}</p>
    </FieldGroup>
  );
}
