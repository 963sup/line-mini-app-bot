import { AppFrame } from "../shell/AppFrame";
import { View } from "../diagnostics/geolocation/View";
export default function Page() {
  return (
    <AppFrame title="定位測試" workflow="location">
      <p className="text-sm text-muted-foreground">位置會傳送至本站驗證，不會儲存。</p>
      <View />
    </AppFrame>
  );
}
