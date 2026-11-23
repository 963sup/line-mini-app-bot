import { useCallback, useEffect, useState, type ReactNode } from "react";
import { ChevronDownIcon, ChevronRightIcon, Folder, Pencil, Plus } from "lucide-react";
import type { components } from "@repo/api-client";
import { useSession } from "@/identity/Session";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { errorMessage, isTemporaryReadError, useLatestRequest } from "@/shared/http/lifecycle";
import { Button } from "@/shared/ui/button";
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from "@/shared/ui/collapsible";
import { Item, ItemGroup, ItemContent, ItemTitle, ItemFooter } from "@/shared/ui/item";

type Resource = components["schemas"]["Object"];
const states = { not_started: "未開始", in_progress: "進行中", completed: "已完成" };
export const typeLabel = (type: string) =>
  !type
    ? "尚未分類"
    : type === "work"
      ? "施作"
      : type === "non_work"
        ? "非施作"
        : `${type}（待確認）`;
export const statusLabel = (item: Resource) =>
  !item.type ? "尚未設定" : item.type === "non_work" ? "不適用" : states[item.status];
export const progressLabel = (item: Resource) =>
  item.type === "non_work"
    ? "不適用"
    : item.type === "work" && item.completedQuantity !== undefined
      ? `${item.completedQuantity}/${item.quantity} ${item.unit}`
      : "待補／待確認";

type Tree = {
  repositoryId: string;
  branch: string;
  activeId: string;
  editor: ReactNode;
  editingBranch: boolean;
  locked: boolean;
  changes: ReadonlyMap<string, Resource>;
  open: (id: string, trigger: HTMLElement, field: string, branch: string) => void;
  addChild: (trigger: HTMLElement, parentId: string, branch: string, expand: () => void) => void;
};

export function ObjectRows({
  items,
  branch,
  tree,
}: {
  items: Resource[];
  branch: string;
  tree: Tree;
}) {
  const [expanded, setExpanded] = useState("");
  return (
    <ItemGroup className="w-full min-w-0 gap-0">
      {items.map((item) => (
        <Collapsible
          key={item.id}
          open={expanded === item.id}
          onOpenChange={(value) => {
            if (!tree.locked && !tree.editingBranch) setExpanded(value ? item.id : "");
          }}
          className="w-full min-w-0"
        >
          <Item role="listitem" className="w-full min-w-0 border-0 px-0 py-0.5">
            <ItemContent className="w-full min-w-0">
              <ItemTitle className="w-full min-w-0 max-w-full">
                <CollapsibleTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="min-w-0 flex-1 justify-start font-normal"
                    disabled={tree.locked || tree.editingBranch}
                    aria-label={`子物件：${item.name}`}
                    title={expanded === item.id ? "收合子物件" : "展開子物件"}
                  >
                    {expanded === item.id ? (
                      <ChevronDownIcon aria-hidden="true" />
                    ) : (
                      <ChevronRightIcon aria-hidden="true" />
                    )}
                    <Folder aria-hidden="true" />
                    <span className="truncate">{item.name}</span>
                  </Button>
                </CollapsibleTrigger>
                <Button
                  variant="link"
                  aria-label={item.name}
                  title={item.canEdit ? `編輯物件：${item.name}` : "物件詳情"}
                  className="h-auto min-w-0 max-w-full justify-start whitespace-normal text-left [overflow-wrap:anywhere]"
                  disabled={tree.locked}
                  onClick={(event) => tree.open(item.id, event.currentTarget, "name", branch)}
                >
                  {item.canEdit ? <Pencil aria-hidden="true" className="size-3.5" /> : "詳情"}
                  <span className="sr-only">名稱：{item.name}</span>
                </Button>
                {branch && item.canEdit && (
                  <Button
                    variant="outline"
                    aria-label={`編輯子物件：${item.name}`}
                    disabled={tree.locked}
                    onClick={(event) => tree.open(item.id, event.currentTarget, "name", branch)}
                  >
                    編輯
                  </Button>
                )}
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  title={`在「${item.name}」下新增子物件`}
                  aria-label={`在「${item.name}」下新增子物件`}
                  disabled={tree.locked}
                  onClick={(event) =>
                    tree.addChild(event.currentTarget, item.id, branch + "/" + item.id, () =>
                      setExpanded(item.id),
                    )
                  }
                >
                  <Plus />
                </Button>
              </ItemTitle>
              <div className="flex flex-wrap items-center gap-x-4 gap-y-1 ps-2 text-xs text-muted-foreground sm:ps-12">
                {(
                  [
                    [
                      "數量",
                      item.quantity && `${item.quantity} ${item.unit ?? ""}`.trim(),
                      "quantity",
                    ],
                    ["類型", item.type && typeLabel(item.type), "type"],
                    ["狀態", item.type && statusLabel(item), ""],
                    [
                      "進度",
                      item.type && progressLabel(item),
                      item.type === "work" ? "completedQuantity" : "",
                    ],
                  ] as const
                ).map(
                  ([label, value, field]) =>
                    value && (
                      <p key={label} className="min-w-0 [overflow-wrap:anywhere]">
                        {item.canEdit && field ? (
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            className="h-auto whitespace-normal px-1 py-1 text-xs"
                            disabled={tree.locked}
                            onClick={(event) =>
                              tree.open(item.id, event.currentTarget, field, branch)
                            }
                          >
                            {label}：{value}
                          </Button>
                        ) : (
                          <>
                            {label}：{value}
                          </>
                        )}
                      </p>
                    ),
                )}
              </div>
            </ItemContent>
            {tree.activeId === item.id && tree.branch === branch && (
              <ItemFooter>{tree.editor}</ItemFooter>
            )}
          </Item>
          <CollapsibleContent className="min-w-0 ps-4 pt-1">
            {expanded === item.id && (
              <ChildList tree={tree} parentId={item.id} branch={branch + "/" + item.id} />
            )}
          </CollapsibleContent>
        </Collapsible>
      ))}
    </ItemGroup>
  );
}

// Each expanded branch owns its page cursor. Writes and the active editor belong
// to the repository view so the same object can appear in both projections.
function ChildList({ parentId, branch, tree }: { parentId: string; branch: string; tree: Tree }) {
  const { request } = useSession();
  const read = useLatestRequest();
  const [items, setItems] = useState<Resource[]>([]);
  const [cursor, setCursor] = useState("");
  const [loading, setLoading] = useState(true);
  const [failedPage, setFailedPage] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [loaded, setLoaded] = useState(false);
  const { repositoryId } = tree;
  const load = useCallback(
    (next = "") => {
      setLoading(true);
      setFailedPage(null);
      setError("");
      void read.run(
        (signal) =>
          request(
            (client, s) =>
              client.GET("/api/v1/repositories/{repositoryId}/objects", {
                signal: s,
                params: { path: { repositoryId }, query: { parentId, cursor: next || undefined } },
              }),
            { signal },
          ),
        (page) => {
          setItems((previous) => (next ? [...previous, ...page.items] : page.items));
          setCursor(page.nextCursor);
          setLoaded(true);
          setLoading(false);
        },
        (reason) => {
          const retryPage = next && isTemporaryReadError(reason) ? next : "";
          if (!retryPage) {
            setItems([]);
            setCursor("");
            setLoaded(false);
          }
          setFailedPage(retryPage);
          setError(errorMessage(reason));
          setLoading(false);
        },
      );
    },
    [read.run, request, repositoryId, parentId],
  );
  useEffect(() => {
    load();
    return read.cancel;
  }, [load, read.cancel]);
  const rows = new Map(items.map((item) => [item.id, item]));
  if (loaded) {
    for (const item of tree.changes.values()) {
      const existing = rows.get(item.id);
      if (item.parentId === parentId && (!existing || existing.version < item.version))
        rows.set(item.id, item);
    }
  }
  const editing = tree.branch === branch && tree.editingBranch;
  return (
    <section aria-label="子物件" className="grid min-w-0 gap-3">
      {editing && !tree.activeId && tree.editor}
      <ReadFeedback
        busy={loading}
        plain={!error}
        label="正在讀取子物件……"
        message={error || (loaded && rows.size === 0 ? "尚無子物件。" : "")}
        onRetry={
          failedPage !== null && !tree.locked && !tree.editingBranch
            ? () => load(failedPage)
            : undefined
        }
      />
      <ObjectRows items={[...rows.values()]} branch={branch} tree={tree} />
      {cursor && failedPage === null && (
        <Button
          disabled={loading || tree.locked || tree.editingBranch}
          onClick={() => load(cursor)}
        >
          載入更多子物件
        </Button>
      )}
    </section>
  );
}
