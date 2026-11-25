import { Shell } from "../../shell/Shell";
import { View } from "../../business/attendance/punch/View";
export default function Page() {
  return (
    <Shell title="打卡" workflow="punch">
      <View />
    </Shell>
  );
}
