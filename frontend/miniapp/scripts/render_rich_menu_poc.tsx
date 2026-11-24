import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { LogIn, LogOut, Repeat2 } from "lucide-react";
import { ImageResponse } from "next/og";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";

import layout from "../../../backend/internal/transport/line/navigation/rich_menu.json";

const icons = {
  "rich-menu-clock-in": LogIn,
  "rich-menu-transfer": Repeat2,
  "rich-menu-clock-out": LogOut,
};

async function main() {
  const [output, fontPath] = process.argv.slice(2);
  if (!output || !fontPath || process.argv.length !== 4) {
    throw new Error("Usage: tsx scripts/render_rich_menu_poc.tsx OUTPUT_DIR TC_FONT_OTF_OR_TTF");
  }
  const [font, css, renderer, lockfile] = await Promise.all([
    readFile(fontPath),
    readFile(new URL("../src/styles/globals.css", import.meta.url), "utf8"),
    readFile(new URL(import.meta.url)),
    readFile(new URL("../../../pnpm-lock.yaml", import.meta.url)),
  ]);
  const root = css.match(/:root\s*\{([^}]+)\}/)?.[1] ?? "";
  function token(name: string) {
    const value = root.match(new RegExp(`--${name}:\\s*([^;]+);`))?.[1].trim();
    assert(value, `Missing shadcn token ${name}`);
    return value;
  }
  const palette = {
    primary: token("primary"),
    primaryForeground: token("primary-foreground"),
    background: token("background"),
    foreground: token("foreground"),
    secondary: token("secondary"),
    border: token("border"),
  };
  assert.equal(layout.pages.length, 3);
  assert.equal(layout.tabWidths.length, layout.pages.length);
  assert.equal(layout.tabWidths.reduce((sum, width) => sum + width, 0), layout.width);
  await mkdir(output, { recursive: true });
  const hash = (data: string | Buffer) => createHash("sha256").update(data).digest("hex");
  const fingerprint = hash(JSON.stringify([
    layout, palette, hash(font), hash(renderer), hash(lockfile), process.version,
  ]));
  const manifestPath = join(output, ".rich-menu-poc-cache.json");
  const filenames = layout.pages.flatMap(({ alias }) => [`${alias}.png`, `${alias}.json`]);
  async function manifest() {
    const digests = await Promise.all(filenames.map(async (name) => hash(await readFile(join(output, name)))));
    return JSON.stringify({ fingerprint, digests });
  }
  const previous = await readFile(manifestPath, "utf8").catch(() => undefined);
  if (previous && previous === await manifest().catch(() => undefined)) {
    console.log("Rich Menu inputs and outputs unchanged. Skipping render.");
    return;
  }

  for (const current of layout.pages) {
    const buttons = ["tab", "action"].flatMap((row) => {
      let x = 0;
      return layout.pages.map((page, index) => {
        const width = layout.tabWidths[index];
        const tab = row === "tab";
        const bounds = {
          x,
          y: tab ? 0 : layout.tabHeight,
          width,
          height: tab ? layout.tabHeight : layout.height - layout.tabHeight,
        };
        x += width;
        const Icon = Object.entries(icons).find(([alias]) => alias === page.alias)?.[1];
        assert(Icon, `Missing Lucide icon for ${page.alias}`);
        return {
          bounds,
          Icon,
          active: tab && current.alias === page.alias,
          action: tab
            ? { type: "richmenuswitch", label: page.label, richMenuAliasId: page.alias, data: `tab=${page.alias}` }
            : { type: "postback", label: page.label, data: page.postback },
          hint: tab ? undefined : page.hint,
        };
      });
    });
    const image = new ImageResponse(
      <div style={{ display: "flex", position: "relative", width: "100%", height: "100%", background: palette.background }}>
        {buttons.map(({ bounds, Icon, action, active, hint }) => (
          <div key={`${action.type}-${action.data}`} style={{
            display: "flex", position: "absolute", left: bounds.x, top: bounds.y,
            width: bounds.width, height: bounds.height, flexDirection: "column",
            alignItems: "center", justifyContent: "center", gap: 32,
            border: `3px solid ${palette.border}`,
            background: active ? palette.primary : hint ? palette.background : palette.secondary,
            color: active ? palette.primaryForeground : palette.foreground,
          }}>
            <img
              alt=""
              width={120}
              height={120}
              src={`data:image/svg+xml;base64,${Buffer.from(renderToStaticMarkup(
                <Icon size={120} color={active ? palette.primaryForeground : palette.primary} />,
              )).toString("base64")}`}
            />
            <span style={{ fontSize: 86 }}>{action.label}</span>
            {hint && <span style={{ fontSize: 42 }}>{hint}</span>}
          </div>
        ))}
      </div>,
      { width: layout.width, height: layout.height, fonts: [{ name: "Menu", data: font }] },
    );
    const png = Buffer.from(await image.arrayBuffer());
    assert.equal(png.subarray(0, 8).toString("hex"), "89504e470d0a1a0a");
    assert.equal(png.readUInt32BE(16), 2500);
    assert.equal(png.readUInt32BE(20), 1686);
    assert(png.length <= 1 << 20, `Image exceeds LINE's 1 MiB limit: ${current.alias}`);
    const request = {
      name: `出勤：${current.label}`,
      chatBarText: layout.chatBarText,
      selected: false,
      size: { width: layout.width, height: layout.height },
      areas: buttons.map(({ bounds, action }) => ({ bounds, action })),
    };
    await writeFile(join(output, `${current.alias}.png`), png);
    await writeFile(join(output, `${current.alias}.json`), JSON.stringify(request, null, 2) + "\n");
    console.log(`${current.alias}: ${png.length} bytes, ${request.areas.length} areas`);
  }
  await writeFile(manifestPath, await manifest());
}

main().catch((error: unknown) => {
  console.error(error);
  process.exitCode = 1;
});
