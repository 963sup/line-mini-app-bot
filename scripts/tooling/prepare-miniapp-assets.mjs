import { cp, mkdir, rm, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";
const root = fileURLToPath(new URL("../../frontend/miniapp/", import.meta.url));
await rm(path.join(root, "public/assets"), { recursive: true, force: true });
await mkdir(path.join(root, "public/assets"), { recursive: true });
// Historical asset URLs remain inert; React owns every workflow.
for (const name of [
  "time.js",
  "api.js",
  "punch.js",
  "history.js",
  "correction.js",
  "geolocation.js",
  "review.js",
  "session.js",
  "miniapp.js",
  "attendance.js",
  "app-api.js",
])
  await writeFile(
    path.join(root, "public/assets", name),
    "// Retired controller. Open the current page to load the React application.\n",
    { mode: 0o644 },
  );

// Self-host the pinned browser PDF worker and its data; no CDN or document URL is used.
const pdfjs = path.join(root, "node_modules/pdfjs-dist");
const pdfAssets = path.join(root, "public/assets/pdfjs");
await mkdir(pdfAssets, { recursive: true });
await cp(path.join(pdfjs, "legacy/build/pdf.worker.min.mjs"), path.join(pdfAssets, "pdf.worker.min.mjs"));
for (const directory of ["cmaps", "standard_fonts", "wasm"]) await cp(path.join(pdfjs, directory), path.join(pdfAssets, directory), { recursive: true });
