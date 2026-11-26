import { View as MissingPunch } from "@/business/attendance/corrections/View";
import { RouteContent } from "@/shell/RouteContent";
import { Shell } from "@/shell/Shell";

export default function Page() {
  return (
    <Shell title="補卡" workflow="missing-punch">
      <RouteContent label="正在開啟補卡……">
        <MissingPunch mode="missing-punch" />
      </RouteContent>
    </Shell>
  );
}
