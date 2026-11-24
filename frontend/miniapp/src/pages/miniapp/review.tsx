import { View as Review } from "@/business/attendance/approvals/View";
import { RouteContent } from "@/shell/RouteContent";
import { Shell } from "@/shell/Shell";

export default function Page() {
  return (
    <Shell title="補卡審核" workflow="review">
      <RouteContent label="正在開啟補卡審核……">
        <Review />
      </RouteContent>
    </Shell>
  );
}
