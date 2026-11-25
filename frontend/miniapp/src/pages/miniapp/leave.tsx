import { View as Leave } from "@/business/attendance/leave/View";
import { RouteContent } from "@/shell/RouteContent";
import { Shell } from "@/shell/Shell";

export default function Page() {
  return (
    <Shell title="請假" workflow="leave">
      <RouteContent label="正在開啟請假……">
        <Leave />
      </RouteContent>
    </Shell>
  );
}
