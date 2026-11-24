import { View as Correction } from "@/business/attendance/corrections/View";
import { RouteContent } from "@/shell/RouteContent";
import { Shell } from "@/shell/Shell";

export default function Page() {
  return (
    <Shell title="申請補卡" workflow="correction">
      <RouteContent label="正在開啟申請補卡……">
        <Correction />
      </RouteContent>
    </Shell>
  );
}
