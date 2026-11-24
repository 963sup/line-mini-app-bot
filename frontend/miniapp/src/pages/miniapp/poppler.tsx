import { Shell } from "@/shell/Shell";
import { Poppler } from "@/business/object/Poppler";

export default function Page() {
  return (
    <Shell title="PDF 擷取" workflow="poppler">
      <Poppler />
    </Shell>
  );
}
