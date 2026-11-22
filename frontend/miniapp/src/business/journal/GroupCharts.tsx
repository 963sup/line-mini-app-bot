import type { ReactNode } from "react";
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { PagePanel } from "@/shared/PagePanel";
import { ChartContainer, ChartLegend, ChartLegendContent, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/shared/ui/chart";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from "@/shared/ui/empty";
import { Table, TableCaption, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/shared/ui/table";
import { Button } from "@/shared/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/shared/ui/collapsible";
import type { components } from "@repo/api-client";

type Journal = components["schemas"]["Journal"];
const config = {
  today: { label: "今日實際" },
  tomorrow: { label: "明日預計" },
} satisfies ChartConfig;

export default function RepositoryChart({ entries, name, renderEntry }: {
  entries: Journal[];
  name: string;
  renderEntry: (entry: Journal) => ReactNode;
}) {
  if (!entries.length) return null;
  const rows = entries.flatMap(journal => journal.progress ?? []);
  return <PagePanel heading={name} description="比較今日實際與明日預計的累計完成度（%）。完整工作名稱與數值可展開進度資料查看。">
    {rows.length ? <>
    <div className="max-h-96 overflow-auto" tabIndex={0} role="region" aria-label={name + "工作進度圖"}>
      <ChartContainer config={config} className="aspect-auto w-full" style={{ height: Math.max(180, rows.length * 28 + 64) }}>
        <BarChart accessibilityLayer layout="vertical" data={rows}>
          <CartesianGrid horizontal={false} />
          <XAxis type="number" domain={[0, 100]} ticks={[0, 25, 50, 75, 100]} />
          <YAxis
            type="category"
            dataKey="name"
            width={88}
            interval={0}
            tickLine={false}
            tickMargin={6}
            axisLine={false}
            tickFormatter={(name: string) => name.length > 6 ? name.slice(0, 6) + "…" : name}
          />
          <ChartTooltip
            cursor={false}
            content={<ChartTooltipContent
              labelFormatter={(_, payload) => payload[0]?.payload?.name}
              formatter={(value, series) => (
                <div className="flex w-full items-center justify-between gap-3">
                  <span>{config[series as keyof typeof config]?.label}</span>
                  <span className="font-mono tabular-nums">{value}%</span>
                </div>
              )}
            />}
          />
          <Bar dataKey="today" fill="var(--primary)" radius={5} barSize={10} isAnimationActive={false} />
          <Bar dataKey="tomorrow" fill="var(--muted-foreground)" radius={5} barSize={10} isAnimationActive={false} />
          <ChartLegend content={<ChartLegendContent />} />
        </BarChart>
      </ChartContainer>
    </div>
    <Collapsible>
      <CollapsibleTrigger asChild><Button type="button" variant="outline">進度資料</Button></CollapsibleTrigger>
      <CollapsibleContent>
        <Table>
          <TableCaption>累計完成度（%）；未提供的數值以「—」顯示。</TableCaption>
          <TableHeader><TableRow><TableHead>工作名稱</TableHead><TableHead>今日實際</TableHead><TableHead>明日預計</TableHead></TableRow></TableHeader>
          <TableBody>{rows.map((row,index) => <TableRow key={index}>
            <TableCell className="min-w-32 whitespace-normal [overflow-wrap:anywhere]">{row.name}</TableCell>
            <TableCell className="tabular-nums">{row.today == null ? "—" : `${row.today}%`}</TableCell>
            <TableCell className="tabular-nums">{row.tomorrow == null ? "—" : `${row.tomorrow}%`}</TableCell>
          </TableRow>)}</TableBody>
        </Table>
      </CollapsibleContent>
    </Collapsible>
    </> : <Empty><EmptyHeader><EmptyTitle>尚無工作進度資料</EmptyTitle><EmptyDescription>仍可展開日誌明細，查看已填寫的內容。</EmptyDescription></EmptyHeader></Empty>}
    <Collapsible>
      <CollapsibleTrigger asChild><Button type="button" variant="outline">日誌明細</Button></CollapsibleTrigger>
      <CollapsibleContent><div className="flex min-w-0 flex-col gap-4 pt-4">{entries.map(renderEntry)}</div></CollapsibleContent>
    </Collapsible>
  </PagePanel>;
}
