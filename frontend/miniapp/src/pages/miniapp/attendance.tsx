import { RouteContent } from "@/shell/RouteContent";
import { useRouter } from "next/router";
import { View as Records } from "@/business/attendance/records/View";
import { View as Leave } from "@/business/attendance/leave/View";
import { View as Correction } from "@/business/attendance/corrections/View";
import { View as Review } from "@/business/attendance/approvals/View";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";
import { Shell } from "@/shell/Shell";

const tabs = [
  { id: "history", label: "出勤紀錄", View: Records },
  { id: "leave", label: "請假紀錄", View: Leave },
  { id: "correction", label: "補卡與更正", View: Correction },
  { id: "review", label: "補卡審核", View: Review },
];

export default function Page() {
  const router = useRouter();
  const selected = tabs.find(tab => tab.id === router.query.tab) ?? tabs[0];

  const applying = selected.id === "correction" && !router.query.record;
  return <Shell title={applying ? "申請補卡" : "出勤服務"} workflow="attendance" shortcut={applying ? "correction" : undefined}>
    <Tabs value={selected.id} onValueChange={tab => {
      if (router.isReady) void router.push({ pathname: "/miniapp/attendance", query: { tab } }, undefined, { shallow: true });
    }}>
      <TabsList aria-label="出勤服務" className="grid w-full group-data-[orientation=horizontal]/tabs:h-auto grid-cols-2 sm:grid-cols-4">
        {tabs.map(({ id, label }) => <TabsTrigger key={id} value={id}>{label}</TabsTrigger>)}
      </TabsList>
      <TabsContent key={selected.id} value={selected.id}>
        <RouteContent label="正在開啟出勤服務……"><selected.View /></RouteContent>
      </TabsContent>
    </Tabs>
  </Shell>;
}
