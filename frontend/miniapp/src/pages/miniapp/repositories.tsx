import { Shell } from "../../shell/Shell";
import { View } from "../../business/repository/View";

export default function Page() {
  return (
    <Shell title="儲存庫" workflow="repositories">
      <View />
    </Shell>
  );
}
