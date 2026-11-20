import { mkdir, rm, writeFile } from "node:fs/promises";
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
