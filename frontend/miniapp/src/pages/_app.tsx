import { MiniAppLayout } from "@/shell/MiniAppProviders";
import type { AppProps } from "next/app";
import "../styles/globals.css";

export default function App({ Component, pageProps, router }: AppProps) {
  return (
    <MiniAppLayout pathname={router.pathname}>
      <Component {...pageProps} />
    </MiniAppLayout>
  );
}
