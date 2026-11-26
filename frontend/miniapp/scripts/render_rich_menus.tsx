import assert from "node:assert/strict";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { Grid3X3, LogIn, LogOut, Repeat2, Smartphone } from "lucide-react";
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
  const [fontPath] = process.argv.slice(2);
  if (!fontPath || process.argv.length !== 3) {
    throw new Error("Usage: pnpm rich-menu:generate TC_FONT_OTF_OR_TTF");
  }
  const output = new URL("../../../assets/line/richmenu/", import.meta.url);
  const [font, css] = await Promise.all([
    readFile(fontPath),
    readFile(new URL("../src/styles/globals.css", import.meta.url), "utf8"),
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
  assert.equal(
    layout.tabWidths.reduce((sum, width) => sum + width, 0),
    layout.width,
  );
  await mkdir(output, { recursive: true });

  for (const current of layout.pages) {
    const Icon = Object.entries(icons).find(([alias]) => alias === current.alias)?.[1];
    assert(Icon, `Missing Lucide icon for ${current.alias}`);
    const footerY = layout.height - layout.footerHeight;
    const image = new ImageResponse(
      <div
        style={{
          display: "flex",
          position: "relative",
          width: "100%",
          height: "100%",
          background: palette.background,
        }}
      >
        {layout.pages.map((page, index) => (
          <div
            key={page.alias}
            style={{
              display: "flex",
              position: "absolute",
              left: layout.tabWidths.slice(0, index).reduce((sum, width) => sum + width, 0),
              top: 0,
              width: layout.tabWidths[index],
              height: layout.tabHeight,
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              border: `3px solid ${palette.border}`,
              background: current.alias === page.alias ? palette.secondary : palette.background,
              color: palette.foreground,
            }}
          >
            <span style={{ fontSize: 72, fontWeight: current.alias === page.alias ? 700 : 500 }}>
              {page.label}
            </span>
          </div>
        ))}
        <div
          style={{
            display: "flex",
            position: "absolute",
            left: 0,
            top: layout.tabHeight,
            width: layout.width,
            height: footerY - layout.tabHeight,
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            gap: 36,
            border: `3px solid ${palette.border}`,
            background: palette.background,
            color: palette.foreground,
          }}
        >
          <img
            alt=""
            width={180}
            height={180}
            src={`data:image/svg+xml;base64,${Buffer.from(
              renderToStaticMarkup(<Icon size={180} color={palette.primary} />),
            ).toString("base64")}`}
          />
          <span style={{ fontSize: 104, fontWeight: 700 }}>{current.label}</span>
          <span style={{ fontSize: 44 }}>{current.hint}</span>
        </div>
        {[
          { label: layout.privateLabel, Icon: Grid3X3 },
          { label: layout.miniAppLabel, Icon: Smartphone },
        ].map((item, index) => (
          <div
            key={item.label}
            style={{
              display: "flex",
              position: "absolute",
              left: index * (layout.width / 2),
              top: footerY,
              width: layout.width / 2,
              height: layout.footerHeight,
              alignItems: "center",
              justifyContent: "center",
              gap: 36,
              border: `3px solid ${palette.border}`,
              background: palette.secondary,
              color: palette.foreground,
            }}
          >
            <img
              alt=""
              width={88}
              height={88}
              src={`data:image/svg+xml;base64,${Buffer.from(
                renderToStaticMarkup(<item.Icon size={88} color={palette.primary} />),
              ).toString("base64")}`}
            />
            <span style={{ fontSize: 58 }}>{item.label}</span>
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
    const filename = current.alias.replaceAll("-", "_") + ".png";
    await writeFile(join(output.pathname, filename), png);
    console.log(`${filename}: ${png.length} bytes`);
  }
}

main().catch((error: unknown) => {
  console.error(error);
  process.exitCode = 1;
});
