import { useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/router";
import { ReadFeedback } from "@/shared/ReadFeedback";

// Static exports and the first browser render must show the same loading state.
export function RouteContent({ label, children }: { label: string; children: ReactNode }) {
  const { isReady } = useRouter();
  const [ready, setReady] = useState(false);
  useEffect(() => { setReady(isReady); }, [isReady]);
  return ready ? children : <ReadFeedback busy label={label} />;
}
