import { RouteContent } from "@/shell/RouteContent";
import { View as Records } from "@/business/attendance/records/View";
import { Shell } from "@/shell/Shell";

export default function Page() {
  return (
    <Shell title="我的出勤" workflow="attendance">
      <RouteContent label="正在開啟出勤紀錄……">
        <Records />
      </RouteContent>
    </Shell>
  );
}
