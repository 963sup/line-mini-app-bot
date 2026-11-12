import React from "react";
import type { ReactNode } from "react";
import { SessionProvider, useSession } from "../identity/Session";
import { useEffect } from "react";
import { RepositoryProvider, useRepositories } from "../business/repository/RepositoryContext";
import { AppFrame } from "./AppFrame";
export function Shell({
  title,
  workflow,
  children,
}: {
  title: string;
  workflow: string;
  children: ReactNode;
}) {
  return (
    <SessionProvider>
      <RepositoryProvider workflow={workflow}>
        <Content title={title} workflow={workflow}>{children}</Content>
      </RepositoryProvider>
    </SessionProvider>
  );
}
function Content({
  title,
  workflow,
  children,
}: {
  title: string;
  workflow: string;
  children: ReactNode;
}) {
  const session = useSession();
  const spaces = useRepositories();
  const { epoch, ready } = session;
  useEffect(() => {
    if (!ready || workflow !== "home") return;
    const params = new URLSearchParams(location.search);
    let leave = params.has("leave"),
      correction = params.get("correction"),
      journal = params.has("journal"),
      repositories = params.has("repositories"),
      repository = params.get("repository") || "";
    try {
      const nested = params.get("liff.state");
      if (nested) {
        const query = new URL(nested, location.origin).searchParams;
        repositories ||= query.has("repositories");
        repository ||= query.get("repository") || "";
        journal ||= query.has("journal");
        leave ||= query.has("leave");
        correction ||= query.get("correction");
      }
    } catch {}
    if (repositories)
      location.replace(
        "/miniapp/repositories" +
          (repository ? "?" + new URLSearchParams({ id: repository }) : ""),
      );
    else if (journal) location.replace("/miniapp/journal");
    else if (correction)
      location.replace(
        "/miniapp/correction?" + new URLSearchParams({ group: correction }),
      );
    else if (leave) location.replace("/miniapp/leave" + location.search);
    else if (params.has("launch") || params.has("liff.state"))
      location.replace("/miniapp/punch" + location.search);
  }, [workflow, ready]);
  return (
    <AppFrame
      title={title}
      workflow={workflow}
      repositoryItem={{
        title: "空間", url: spaces.scope ? spaces.href(spaces.scope) : "/miniapp/repositories", isActive: workflow === "repositories", onExpand: spaces.ensure,
        items: [
          ...spaces.items.map((item) => ({ title: item.name || "群組空間", url: spaces.href(item.id), isActive: item.id === spaces.selectedID })),
          { title: spaces.status, status: true },
          ...(spaces.cursor ? [{ title: "載入更多空間", onSelect: spaces.loadMore, disabled: spaces.busy, keepOpen: true }] : []),
          { title: spaces.items.length ? "重新載入空間清單" : "載入空間", onSelect: spaces.reload, disabled: !spaces.canLoad || spaces.busy, keepOpen: true },
        ],
      }}
      account={{
        name: session.user?.displayName || "LINE 身分",
        detail: session.user?.userId,
        avatar: session.pictureUrl,
        status: session.status,
        actions: [
          ...(session.ready &&
          !session.user &&
          !session.inClient &&
          !session.busy
            ? [
                {
                  id: "login",
                  label: "使用 LINE 登入",
                  onSelect: session.login,
                },
              ]
            : []),
          {
            id: "verify",
            label:
              session.ready || session.busy ? "重新驗證身分" : "重試初始化",
            onSelect: () => void session.verify(),
            disabled: session.busy,
          },
          ...(session.canLogout
            ? [
                {
                  id: "logout",
                  label: "登出",
                  onSelect: () => void session.logout(),
                  disabled: session.busy,
                },
              ]
            : []),
        ],
      }}
    >
      <div
        className="flex flex-col gap-3"
        key={workflow === "journal" ? "journal" : epoch}
      >
        {children}
      </div>
    </AppFrame>
  );
}
