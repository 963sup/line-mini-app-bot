import { PagePanel } from "./PagePanel";
import { TabsContent, TabsTrigger } from "./ui/tabs";

export function AuditTab() {
  return <TabsTrigger value="audit">稽核</TabsTrigger>;
}

export function AuditPanel({ scope }: { scope: string }) {
  return <TabsContent value="audit">
    <PagePanel heading="稽核"><p>{scope}的修改歷史尚未開放查詢。</p></PagePanel>
  </TabsContent>;
}
