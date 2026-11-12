import { Progress } from "@/shared/ui/progress";
import React, { useState } from "react";
import { AppFrame } from "@/shell/AppFrame";
import {
  LayoutDashboard,
  CircleDot,
  MessagesSquare,
  Tags,
  Flag,
  History,
  Layers,
  Plus,
  ArrowUpRight,
  Search,
  ChevronRight,
} from "lucide-react";
import { Button } from "@/shared/ui/button";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/shared/ui/card";
import { Badge } from "@/shared/ui/badge";
import { Input } from "@/shared/ui/input";
import { Field, FieldGroup, FieldLabel } from "@/shared/ui/field";
import { Textarea } from "@/shared/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/shared/ui/dialog";
import {
  Empty,
  EmptyHeader,
  EmptyTitle,
  EmptyDescription,
} from "@/shared/ui/empty";

const navigation = [
  { title: "總覽", icon: LayoutDashboard },
  { title: "事項", icon: CircleDot },
  { title: "討論", icon: MessagesSquare },
  { title: "標籤", icon: Tags },
  { title: "里程碑", icon: Flag },
  { title: "專案", icon: Layers },
  { title: "稽核紀錄", icon: History },
];
type Item = {
  number: number;
  title: string;
  state: string;
  label: string;
  owner: string;
  detail: string;
};
const seed: Item[] = [
  {
    number: 12,
    title: "確認一樓材料到場與點收",
    state: "進行中",
    label: "採購",
    owner: "林先生",
    detail: "核對材料清單、數量與規格，完成後回覆點收結果。",
  },
  {
    number: 10,
    title: "更新第二階段施工圖說",
    state: "待處理",
    label: "設計",
    owner: "陳小姐",
    detail: "整理現場回饋並更新圖說，將需確認的項目列入討論。",
  },
  {
    number: 8,
    title: "完成現場安全檢查",
    state: "已完成",
    label: "現場",
    owner: "王先生",
    detail: "檢查施工動線、防護設施及作業區域。",
  },
];
const discussions = [
  "第二階段施工順序與介面協調",
  "材料進場前需要確認哪些文件？",
  "每週現場會議紀錄整理方式",
];

export default function Workspace() {
  const [page, setPage] = useState("總覽");
  const [items, setItems] = useState(seed);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState("全部");
  const [selected, setSelected] = useState<Item | null>(null);
  const [topic, setTopic] = useState<string | null>(null);
  const [topics, setTopics] = useState(discussions);
  const [topicDetails, setTopicDetails] = useState<Record<string, string>>({});
  const [creating, setCreating] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [notice, setNotice] = useState("");
  const [comment, setComment] = useState("");
  const [comments, setComments] = useState<Record<string, string[]>>({});
  function navigate(value: string) {
    setPage(value);
    setQuery("");
    setFilter("全部");
  }
  const visible = items.filter(
    (item) =>
      item.title.includes(query) &&
      (filter === "全部" || item.state === filter),
  );
  const commentKey = selected
    ? `issue-${selected.number}`
    : `discussion-${topic}`;
  const rows = (data: Item[]) => (
    <div className="flex flex-col">
      {data.map((item) => (
        <button
          className="flex w-full flex-wrap items-center gap-3 border-b py-5 text-left hover:text-primary [&>div]:min-w-0 [&>div]:flex-1 [&>div]:basis-[55%] sm:[&>div]:basis-auto [&_strong]:text-sm"
          key={item.number}
          onClick={() => {
            setSelected(item);
            setComment("");
          }}
        >
          <CircleDot size={18} aria-hidden="true" />
          <div>
            <strong>{item.title}</strong>
            <div className="text-sm text-muted-foreground">
              #{item.number} · {item.owner} · {item.label}
            </div>
          </div>
          <Badge variant={item.state === "已完成" ? "secondary" : "outline"}>
            {item.state}
          </Badge>
          <ChevronRight size={16} aria-hidden="true" />
        </button>
      ))}
    </div>
  );
  return (
    <AppFrame
      title="空間 UI 預覽"
      workflow="preview"
      wide
      contextItems={navigation.map(({ title, icon }) => ({
        label: title,
        icon,
        active: page === title,
        onSelect: () => navigate(title),
      }))}
    >
      <div className="mb-4 flex items-start justify-between gap-3 sm:items-center [&_h2]:mb-2 [&_h2]:text-2xl [&_h2]:font-bold [&_h2]:tracking-tight sm:[&_h2]:text-[34px]">
        <div>
          <p className="mb-2.5 text-[11px] font-bold tracking-widest text-primary">
            WORKSPACE / {page === "總覽" ? "OVERVIEW" : "COLLABORATION"}
          </p>
          <h2>{page === "總覽" ? "讓每件工作，都有進展。" : page}</h2>
          <p className="text-sm text-muted-foreground">
            {page === "總覽"
              ? "在同一個空間掌握事項、交流想法與追蹤目標。"
              : "工程 A 的共同工作紀錄"}
          </p>
        </div>
        {["事項", "討論"].includes(page) && (
          <Button
            onClick={() => {
              setTitle("");
              setDescription("");
              setCreating(true);
            }}
          >
            <Plus size={16} />
            {page === "事項" ? "新增事項" : "發起討論"}
          </Button>
        )}
      </div>
      <p className="mb-6 text-xs text-muted-foreground">
        目前為介面預覽，所有內容皆為示範資料。
      </p>
      <div role="status" aria-live="polite">
        {notice}
      </div>
      {page === "總覽" && (
        <>
          <div className="mb-6 grid grid-cols-3 gap-2 sm:gap-4 [&_[data-slot=card-header]]:px-3 sm:[&_[data-slot=card-header]]:px-6 [&_[data-slot=card-content]]:hidden sm:[&_[data-slot=card-content]]:flex [&_[data-slot=card-description]]:text-xs">
            {[
              [
                "未完成事項",
                items.filter((i) => i.state !== "已完成").length,
                "讓待辦持續往前",
              ],
              ["討論主題", topics.length, "交流與共同決定"],
              ["階段里程碑", 2, "分階段完成目標"],
            ].map(([name, count, hint]) => (
              <Card key={name}>
                <CardHeader>
                  <CardDescription>{name}</CardDescription>
                  <CardTitle className="text-[28px] tabular-nums sm:text-[38px]">
                    {count}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <span className="text-sm text-muted-foreground">{hint}</span>
                </CardContent>
              </Card>
            ))}
          </div>
          <div className="grid grid-cols-1 gap-5 min-[900px]:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)]">
            <Card>
              <CardHeader>
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <CardTitle>接下來要做的事</CardTitle>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => navigate("事項")}
                  >
                    全部事項 <ArrowUpRight size={14} />
                  </Button>
                </div>
                <CardDescription>保持焦點，從下一件事開始。</CardDescription>
              </CardHeader>
              <CardContent>
                {rows(items.filter((i) => i.state !== "已完成"))}
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>近期討論</CardTitle>
                <CardDescription>讓資訊留在團隊的共同空間。</CardDescription>
              </CardHeader>
              <CardContent>
                {topics.slice(0, 3).map((t, index) => (
                  <button
                    className="flex w-full items-center gap-3 border-b py-4 text-left text-sm hover:text-primary [&>span]:flex-1 [&_small]:mt-1 [&_small]:block"
                    key={t + index}
                    onClick={() => {
                      setTopic(t);
                      setComment("");
                    }}
                  >
                    <MessagesSquare size={18} />
                    <span>{t}</span>
                    <ChevronRight size={15} />
                  </button>
                ))}
              </CardContent>
            </Card>
          </div>
        </>
      )}
      {page === "事項" && (
        <Card>
          <CardHeader>
            <div className="flex max-w-[460px] items-center gap-3">
              <Search size={18} />
              <Input
                aria-label="搜尋事項"
                placeholder="搜尋事項名稱…"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
              />
            </div>
            <div className="flex flex-wrap gap-2">
              {["全部", "待處理", "進行中", "已完成"].map((s) => (
                <Button
                  key={s}
                  size="sm"
                  variant={filter === s ? "secondary" : "ghost"}
                  aria-pressed={filter === s}
                  onClick={() => setFilter(s)}
                >
                  {s}
                </Button>
              ))}
            </div>
          </CardHeader>
          <CardContent>
            {visible.length ? (
              rows(visible)
            ) : (
              <Empty>
                <EmptyHeader>
                  <EmptyTitle>沒有符合的事項</EmptyTitle>
                  <EmptyDescription>試試其他關鍵字或狀態。</EmptyDescription>
                </EmptyHeader>
              </Empty>
            )}
          </CardContent>
        </Card>
      )}
      {page === "討論" && (
        <Card>
          <CardHeader>
            <CardTitle>團隊正在聊什麼</CardTitle>
            <CardDescription>
              提案、問答與經驗，都可以從這裡開始。
            </CardDescription>
          </CardHeader>
          <CardContent>
            {topics.map((t, index) => (
              <button
                className="flex w-full items-center gap-3 border-b py-4 text-left text-sm hover:text-primary [&>span]:flex-1 [&_small]:mt-1 [&_small]:block"
                key={t + index}
                onClick={() => {
                  setTopic(t);
                  setComment("");
                }}
              >
                <MessagesSquare size={20} />
                <span>
                  <strong>{t}</strong>
                  <small className="text-sm text-muted-foreground">
                    {index === 1 ? "問答" : "一般討論"} · 示範主題
                  </small>
                </span>
                <ChevronRight size={16} />
              </button>
            ))}
          </CardContent>
        </Card>
      )}
      {page === "標籤" && (
        <Card>
          <CardHeader>
            <CardTitle>共用標籤</CardTitle>
            <CardDescription>事項與討論使用同一套空間分類。</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {["採購", "設計", "現場", "待確認", "安全"].map((t) => (
                <Badge variant="secondary" key={t}>
                  {t}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
      {page === "里程碑" && (
        <div className="grid grid-cols-1 gap-5 min-[900px]:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)]">
          {["第一階段 · 現場準備", "第二階段 · 施工協調"].map((t, i) => (
            <Card key={t}>
              <CardHeader>
                <Flag size={22} />
                <CardTitle>{t}</CardTitle>
                <CardDescription>
                  {i ? "確認施工圖說及各工種介面" : "完成進場、點收與安全檢查"}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p>示範完成度 {i ? "25" : "75"}%</p>
                <Progress value={i ? 25 : 75} aria-label={t + "完成度"} />
              </CardContent>
            </Card>
          ))}
        </div>
      )}
      {page === "專案" && (
        <Card>
          <CardHeader>
            <CardTitle>施工工作看板</CardTitle>
            <CardDescription>
              同一批事項的不同視圖；本頁僅預覽狀態分欄。
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3 [&>section]:rounded-xl [&>section]:bg-muted [&>section]:p-4 [&_h3]:mb-3 [&_h3]:text-sm [&_h3]:font-semibold">
              {["待處理", "進行中", "已完成"].map((s) => (
                <section key={s}>
                  <h3 className="font-semibold">{s}</h3>
                  {items
                    .filter((i) => i.state === s)
                    .map((i) => (
                      <button
                        className="w-full rounded-lg border bg-card p-4 text-left text-card-foreground hover:bg-accent"
                        key={i.number}
                        onClick={() => setSelected(i)}
                      >
                        #{i.number} {i.title}
                      </button>
                    ))}
                </section>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
      {page === "稽核紀錄" && (
        <Card>
          <CardHeader>
            <CardTitle>重要操作紀錄</CardTitle>
            <CardDescription>
              以下為示範紀錄，並非正式稽核證據。
            </CardDescription>
          </CardHeader>
          <CardContent>
            {[
              "林先生將事項 #12 更新為進行中",
              "陳小姐建立事項 #10",
              "王先生完成事項 #8",
            ].map((t, i) => (
              <div className="flex items-start gap-3 py-4" key={t}>
                <History size={18} />
                <div>
                  <strong>{t}</strong>
                  <p className="text-sm text-muted-foreground">
                    示範時間 · {i + 1} 小時前
                  </p>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
      <Dialog open={creating} onOpenChange={setCreating}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {page === "討論" ? "發起討論" : "新增事項"}
            </DialogTitle>
            <DialogDescription>
              示範操作，重新整理頁面後即會還原。
            </DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (!title.trim()) return;
              if (page === "討論") {
                setTopics([title.trim(), ...topics]);
                setTopicDetails({
                  ...topicDetails,
                  [title.trim()]: description,
                });
              } else
                setItems([
                  ...items,
                  {
                    number: Math.max(...items.map((i) => i.number)) + 1,
                    title: title.trim(),
                    detail: description,
                    state: "待處理",
                    owner: "尚未指派",
                    label: "未分類",
                  },
                ]);
              setCreating(false);
              setNotice("已新增至本次預覽。");
            }}
          >
            <FieldGroup>
              <Field>
                <FieldLabel
                  className="w-full flex-col items-stretch gap-2 whitespace-normal"
                  htmlFor="preview-title"
                >
                  標題
                </FieldLabel>
                <Input
                  id="preview-title"
                  required
                  maxLength={120}
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                />
              </Field>
              <Field>
                <FieldLabel
                  className="w-full flex-col items-stretch gap-2 whitespace-normal"
                  htmlFor="preview-description"
                >
                  內容
                </FieldLabel>
                <Textarea
                  id="preview-description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </Field>
            </FieldGroup>
            <DialogFooter className="mt-6">
              <Button type="submit">加入預覽</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog
        open={!!selected || !!topic}
        onOpenChange={(open) => {
          if (!open) {
            setSelected(null);
            setTopic(null);
          }
        }}
      >
        <DialogContent className="max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {selected ? `#${selected.number} ${selected.title}` : topic}
            </DialogTitle>
            <DialogDescription>
              {selected
                ? `${selected.owner} · ${selected.label} · ${selected.state}`
                : "示範討論內容"}
            </DialogDescription>
          </DialogHeader>
          <p>
            {selected?.detail ||
              (topic && topicDetails[topic]) ||
              "在這裡整理想法、補充現場資訊，一起確認下一步。"}
          </p>
          {selected && (
            <Button
              variant="outline"
              onClick={() => {
                const changed = {
                  ...selected,
                  state: selected.state === "已完成" ? "待處理" : "已完成",
                };
                setItems(
                  items.map((i) => (i.number === changed.number ? changed : i)),
                );
                setSelected(changed);
              }}
            >
              {selected.state === "已完成"
                ? "重新開啟（示範）"
                : "標記完成（示範）"}
            </Button>
          )}
          <h3 className="font-semibold">留言</h3>
          {(comments[commentKey] || []).map((c, i) => (
            <p key={i}>{c}</p>
          ))}
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (!comment.trim()) return;
              setComments({
                ...comments,
                [commentKey]: [...(comments[commentKey] || []), comment.trim()],
              });
              setComment("");
            }}
          >
            <FieldGroup>
              <Field>
                <FieldLabel
                  className="w-full flex-col items-stretch gap-2 whitespace-normal"
                  htmlFor="preview-comment"
                >
                  新增留言
                </FieldLabel>
                <Textarea
                  id="preview-comment"
                  required
                  value={comment}
                  onChange={(e) => setComment(e.target.value)}
                />
              </Field>
            </FieldGroup>
            <DialogFooter className="mt-4">
              <Button type="submit">送出示範留言</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </AppFrame>
  );
}
