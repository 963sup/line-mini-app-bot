import { Shell } from "@/shell/Shell";
import { Poppler } from "@/business/operations/Poppler";

export default function Page() {
  return (
    <Shell title="PDF 文字擷取" workflow="poppler">
      <Poppler />
    </Shell>
  );
}
