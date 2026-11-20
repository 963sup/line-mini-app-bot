import React, { useRef, useState } from "react";
import { useSession } from "@/identity/Session";
import { FormField } from "@/shared/FormField";
import { PagePanel } from "@/shared/PagePanel";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Status } from "@/shared/Status";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { errorMessage, useLatestRequest } from "@/shared/http/lifecycle";
import { useRepositoryGuard } from "./NavigationGuard";
import { useSpaces } from "./useSpaces";

export function Spaces({ repositoryId }: { repositoryId: string }) {
  const { user, request } = useSession();
  const { spaces, setSpaces, spacesBusy, spaceStatus, spaceCursor, loadSpaces } = useSpaces(repositoryId);
  const [name, setName] = useState(""), [message, setMessage] = useState(""), [busy, setBusy] = useState(false);
  const input = useRef<HTMLInputElement>(null), lock = useRef(false);
  const pending = useRef<{ key: string; name: string } | null>(null);
  const guard = useRepositoryGuard(!!name.trim(), () => input.current?.focus());
  const write = useLatestRequest();
  function create() {
    if (!user || !name.trim() || lock.current || spacesBusy || !guard.canWrite()) return;
    const retry = !!pending.current;
    const attempt = pending.current ?? { key: crypto.randomUUID(), name: name.trim() };
    pending.current = attempt;
    lock.current = true; guard.block(true); setBusy(true); setMessage("");
    void write.run(signal => request((client, s) => client.POST("/api/v1/repositories/{repositoryId}/spaces", {
      signal: s, params: { path: { repositoryId }, header: { "Idempotency-Key": attempt.key } }, body: { name: attempt.name }
    }), { signal }), item => {
      setSpaces(previous => [...previous.filter(space => space.id !== item.id), item]);
      pending.current = null; lock.current = false; guard.block(false); setBusy(false); setName(""); setMessage("已建立空間。");
    }, error => {
      if (!retry && ["invalid_input", "forbidden", "not_found", "idempotency_conflict"].includes((error as { code?: string }).code || "")) {
        pending.current = null; guard.block(false);
      }
      lock.current = false; setBusy(false); setMessage(errorMessage(error));
    });
  }
  return <PagePanel heading="空間">
    <form className="grid gap-2" onSubmit={event => { event.preventDefault(); create(); }}>
      <FormField>空間名稱<Input ref={input} required maxLength={120} disabled={!!pending.current} value={name} onChange={event => setName(event.target.value)} /></FormField>
      <Button type="submit" disabled={!user || busy || spacesBusy || !name.trim()}>{pending.current && !busy ? "重試原空間建立請求" : "新增空間"}</Button>
    </form>
    <Status role="status">{message}</Status>
    <ReadFeedback busy={spacesBusy} message={spaceStatus} label="正在讀取空間……" />
    {spaces.length > 0 ? <ul>{spaces.map(space => <li className="break-words" key={space.id}>{space.name}</li>)}</ul> : !spacesBusy && !spaceStatus && <p>尚無空間</p>}
    {spaceStatus && <Button variant="outline" disabled={busy} onClick={() => loadSpaces()}>重試讀取</Button>}
    {spaceCursor && <Button variant="outline" disabled={busy || spacesBusy} onClick={() => loadSpaces(spaceCursor)}>載入更多空間</Button>}
  </PagePanel>;
}
