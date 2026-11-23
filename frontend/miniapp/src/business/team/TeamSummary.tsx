import type { components } from "@repo/api-client";
import { AuditPanel, AuditTab } from "@/shared/AuditTab";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";

type Props = {
  name: string;
  management: string;
  auditStatus?: components["schemas"]["AuditHistoryStatus"];
};

export function TeamSummary({ name, management, auditStatus }: Props) {
  return (
    <details>
      <summary className="cursor-pointer break-words py-2">{name}</summary>
      <Tabs defaultValue="overview">
        <TabsList aria-label={`${name}團隊功能`}>
          <TabsTrigger value="overview">團隊資料</TabsTrigger>
          {auditStatus === "unavailable" && <AuditTab />}
        </TabsList>
        <TabsContent value="overview">{management}</TabsContent>
        {auditStatus === "unavailable" && <AuditPanel scope={name} />}
      </Tabs>
    </details>
  );
}
