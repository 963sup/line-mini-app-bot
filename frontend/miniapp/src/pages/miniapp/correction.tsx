import { View as Correction } from "@/business/attendance/corrections/View";
import { View as Records } from "@/business/attendance/records/View";
import { RouteContent } from "@/shell/RouteContent";
import { Shell } from "@/shell/Shell";
import { useRouter } from "next/router";

export default function Page() {
  const selected = typeof useRouter().query.record === "string";
  return (
    <Shell title="更正" workflow="correction">
      <RouteContent label="正在開啟更正……">
        {selected ? <Correction mode="correction" /> : <Records />}
      </RouteContent>
    </Shell>
  );
}
