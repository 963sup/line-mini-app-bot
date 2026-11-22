import { AppFrame } from "../shell/AppFrame";
import { View } from "../diagnostics/geolocation/View";
export default function Page() {
  return (
    <AppFrame title="定位測試" workflow="location">
      <View />
    </AppFrame>
  );
}
