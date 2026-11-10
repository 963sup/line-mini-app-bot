import type { NextConfig } from "next";
import { PHASE_DEVELOPMENT_SERVER } from "next/constants";
export default function config(phase: string): NextConfig {
  const development = phase === PHASE_DEVELOPMENT_SERVER;
  return {
    ...(development ? {} : { output: "export" as const }),
    poweredByHeader: false,
    reactStrictMode: true,
    images: { unoptimized: true },
    ...(development
      ? {
          async rewrites() {
            return [
              {
                source: "/api/:path*",
                destination: "http://localhost:8080/api/:path*",
              },
            ];
          },
        }
      : {}),
  };
}
