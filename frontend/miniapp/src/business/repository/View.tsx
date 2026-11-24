import dynamic from "next/dynamic";
import { useEffect, useState } from "react";
import { Empty, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { FormField } from "@/shared/FormField";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native_select";
import { PagePanel } from "@/shared/PagePanel";
import { Button } from "@/shared/ui/button";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { useSession } from "@/identity/Session";
import { RepositoryNavigationContext, useRepositoryNavigation } from "./NavigationGuard";
import { QuantityHistory } from "../object/QuantityHistory";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";
import { useRepositories } from "./RepositoryContext";
import { navigation } from "./navigation";
import { displayTime } from "../../shared/time/time";
const Objects = dynamic(() => import("../object/View").then((module) => module.View), {
  loading: Loading,
});
const Issues = dynamic(() => import("../issue/View").then((module) => module.View), {
  loading: Loading,
});
const Discussions = dynamic(() => import("../discussion/View").then((module) => module.View), {
  loading: Loading,
});
const Milestones = dynamic(() => import("./milestones/View").then((module) => module.View), {
  loading: Loading,
});
const Insights = dynamic(() => import("./insights/View").then((module) => module.View), {
  loading: Loading,
});
const Wiki = dynamic(() => import("./wiki/View").then((module) => module.View), {
  loading: Loading,
});
const Labels = dynamic(() => import("./labels/View").then((module) => module.View), {
  loading: Loading,
});

const Spaces = dynamic(() => import("./Spaces").then((module) => module.Spaces), {
  loading: Loading,
});

function Loading() {
  return <ReadFeedback busy label="正在載入功能……" />;
}

export function View() {
  const { user, epoch } = useSession();
  const { selectedID } = useRepositories();
  return <Content key={`${selectedID}:${user?.userId}:${epoch}`} />;
}

function Content() {
  const repositories = useRepositories();
  const { selected, selectedID, detailBusy, detailStatus, reloadDetail, canLoad } = repositories;
  const [wikiId] = useState(() => {
    if (typeof window === "undefined") return "";
    const params = new URLSearchParams(location.search);
    try {
      return (
        params.get("wiki") ||
        new URL(params.get("liff.state") || "/", location.origin).searchParams.get("wiki") ||
        ""
      );
    } catch {
      return "";
    }
  });
  const [objectId] = useState(() => {
    if (typeof window === "undefined") return "";
    const query = new URLSearchParams(location.search);
    return query.get("id") === selectedID ? query.get("object") || "" : "";
  });
  const [page, setPage] = useState(wikiId ? "Wiki" : objectId ? "物件" : "總覽");
  const guard = useRepositoryNavigation();
  useEffect(() => {
    if (!detailBusy) guard.settled();
  }, [page, detailBusy]);
  return (
    <RepositoryNavigationContext.Provider value={guard.context}>
      <section className="flex min-w-0 flex-col gap-4" aria-label="空間">
        {selectedID ? (
          <>
            {!selected && canLoad && !detailBusy && detailStatus && (
              <Button
                disabled={!canLoad || detailBusy || guard.blocked}
                onClick={() => guard.navigate(reloadDetail)}
              >
                重新載入空間
              </Button>
            )}
            <ReadFeedback busy={detailBusy} message={detailStatus} label="正在讀取空間……" />
          </>
        ) : (
          <PagePanel heading="選擇儲存庫">
            {repositories.items.length > 0 && (
              <FormField>
                儲存庫
                <NativeSelect
                  value=""
                  onChange={(event) => {
                    if (event.target.value)
                      window.location.assign(repositories.href(event.target.value));
                  }}
                >
                  <NativeSelectOption value="">選擇要開啟的儲存庫</NativeSelectOption>
                  {repositories.items.map((item) => (
                    <NativeSelectOption key={item.id} value={item.id}>
                      {item.name || "群組空間"}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
              </FormField>
            )}
            <ReadFeedback
              busy={repositories.busy}
              label="正在讀取儲存庫……"
              message={repositories.catalogFailed || !canLoad ? repositories.status : undefined}
              onRetry={repositories.catalogFailed ? repositories.retryCatalog : undefined}
            />
            {canLoad &&
              !repositories.busy &&
              !repositories.catalogFailed &&
              !repositories.items.length &&
              repositories.status && (
                <Empty>
                  <EmptyHeader>
                    <EmptyTitle>{repositories.status}</EmptyTitle>
                  </EmptyHeader>
                </Empty>
              )}
            {repositories.cursor && (
              <Button
                type="button"
                variant="outline"
                disabled={repositories.busy}
                onClick={repositories.loadMore}
              >
                載入更多儲存庫
              </Button>
            )}
          </PagePanel>
        )}
        <noscript>請啟用 JavaScript 並完成 LINE 身分驗證，以載入側邊欄空間及內容。</noscript>
        {selected && (
          <>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2 className="min-w-0 text-2xl font-semibold tracking-tight [overflow-wrap:anywhere]">
                {selected.name || "儲存庫"}
              </h2>
            </div>
            <Tabs
              value={page}
              activationMode="manual"
              onValueChange={(value) => guard.navigate(() => setPage(value))}
            >
              <TabsList
                aria-label="空間功能"
                className="max-w-full flex-wrap justify-start group-data-[orientation=horizontal]/tabs:h-auto"
              >
                {navigation.map(({ title, icon: Icon }) => (
                  <TabsTrigger
                    key={title}
                    value={title}
                    className="h-auto"
                    aria-label={title}
                    disabled={guard.blocked && title !== page}
                  >
                    <Icon aria-hidden="true" />
                    {title}
                  </TabsTrigger>
                ))}
              </TabsList>
              <TabsContent value="總覽">
                <div className="grid min-w-0 gap-4 lg:grid-cols-2">
                  <PagePanel heading="空間資料">
                    <dl className="grid gap-2 text-sm [&>dt]:font-medium [&>dd]:break-all [&>dd]:text-muted-foreground">
                      <dt>儲存庫 ID</dt>
                      <dd>{selected.id}</dd>
                      <dt>擁有者 ID</dt>
                      <dd>{selected.ownerId}</dd>
                      <dt>狀態</dt>
                      <dd>{selected.status === "ACTIVE" ? "啟用" : "停用"}</dd>
                      <dt>版本</dt>
                      <dd>{selected.version}</dd>
                      <dt>建立時間</dt>
                      <dd>{displayTime(selected.createdAt)}</dd>
                      <dt>更新時間</dt>
                      <dd>{displayTime(selected.updatedAt)}</dd>
                    </dl>
                  </PagePanel>
                  <PagePanel heading="動態">
                    <div className="flex flex-wrap gap-2">
                      <Button asChild>
                        <a
                          href={
                            "/miniapp/feed?" +
                            new URLSearchParams({ repositoryId: selected.id, compose: "1" })
                          }
                        >
                          寫日誌
                        </a>
                      </Button>
                      <Button asChild variant="outline">
                        <a
                          href={
                            "/miniapp/feed?" + new URLSearchParams({ repositoryId: selected.id })
                          }
                        >
                          查看日誌
                        </a>
                      </Button>
                    </div>
                  </PagePanel>
                  {selected.groupId && (
                    <PagePanel heading="出勤與群組服務">
                      <a
                        className="break-words text-primary underline underline-offset-4 focus-visible:outline-2 focus-visible:outline-ring"
                        href={
                          "/miniapp/attendance?" + new URLSearchParams({ group: selected.groupId })
                        }
                      >
                        查看本人出勤
                      </a>
                    </PagePanel>
                  )}
                </div>
              </TabsContent>
              <TabsContent value="空間">
                <Spaces repositoryId={selected.id} />
              </TabsContent>
              <TabsContent value="物件">
                <Objects repositoryId={selected.id} initialObjectId={objectId} />
              </TabsContent>
              <TabsContent value="問題">
                <Issues repositoryId={selected.id} />
              </TabsContent>
              <TabsContent value="討論">
                <Discussions repositoryId={selected.id} />
              </TabsContent>
              <TabsContent value="標籤">
                <Labels repositoryId={selected.id} />
              </TabsContent>
              <TabsContent value="里程碑">
                <Milestones repositoryId={selected.id} />
              </TabsContent>
              <TabsContent value="Wiki">
                <Wiki repositoryId={selected.id} initialId={wikiId} />
              </TabsContent>
              <TabsContent value="洞察">
                <Insights repositoryId={selected.id} />
              </TabsContent>
              <TabsContent value="稽核">
                <QuantityHistory key={selected.id} repositoryId={selected.id} />
              </TabsContent>
            </Tabs>
          </>
        )}
      </section>
      {guard.dialog}
    </RepositoryNavigationContext.Provider>
  );
}
