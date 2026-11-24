import { View } from "@/business/explore/View";
import { Shell } from "@/shell/Shell";

export default function Page() {
  return (
    <Shell title="Explore" workflow="explore">
      <View />
    </Shell>
  );
}
