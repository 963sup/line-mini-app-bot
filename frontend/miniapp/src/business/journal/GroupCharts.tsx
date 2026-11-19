import type { ReactNode } from "react";
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { PagePanel } from "@/shared/PagePanel";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/shared/ui/chart";
import { Button } from "@/shared/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/shared/ui/collapsible";
import type { components } from "@repo/api-client";

type Journal = components["schemas"]["Journal"];
const config = {
  today: { label: "今日實際" },
  tomorrow: { label: "明日預計" },
} satisfies ChartConfig;

export default function GroupCharts({ entries, groups, renderEntry }: {
  entries: Journal[];
  groups: { id: string; name: string }[];
  renderEntry: (entry: Journal) => ReactNode;
}) {
  const byGroup = new Map<string, Journal[]>();
  for (const entry of entries) {
    const journals = byGroup.get(entry.groupId);
    if (journals) journals.push(entry);
    else byGroup.set(entry.groupId, [entry]);
  }
  return Array.from(byGroup, ([id, journals]) => {
    const name = groups.find(group => group.id === id)?.name || id;
    const rows = journals.flatMap(journal => [
      ...journal.todayItems.map(item => ({ name: item.name, today: item.completionPercent })),
      ...journal.tomorrowItems.map(item => ({ name: item.name, tomorrow: item.expectedCompletionPercent })),
    ]);
    return <PagePanel key={id} heading={name}>
      <p className="flex flex-wrap gap-x-4">
        <span><span aria-hidden="true" className="mr-1 inline-block size-2 bg-primary" />今日實際（%）</span>
        <span><span aria-hidden="true" className="mr-1 inline-block size-2 bg-muted-foreground" />明日預計（%）</span>
      </p>
      <div className="max-h-96 overflow-auto" tabIndex={0} role="region" aria-label={name + "工作進度圖"}>
        <ChartContainer config={config} className="aspect-auto w-full" style={{ height: Math.max(160, rows.length * 28 + 32) }}>
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
          </BarChart>
        </ChartContainer>
      </div>
      <Collapsible>
        <CollapsibleTrigger asChild><Button type="button" variant="outline">日誌明細</Button></CollapsibleTrigger>
        <CollapsibleContent>{journals.map(renderEntry)}</CollapsibleContent>
      </Collapsible>
    </PagePanel>;
  });
}
