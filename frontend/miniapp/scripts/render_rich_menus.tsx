import assert from "node:assert/strict";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";
import {
  Archive,
  CalendarDays,
  CalendarPlus,
  ClipboardCheck,
  Clock3,
  Grid3X3,
  LogIn,
  LogOut,
  Repeat2,
  Smartphone,
  type LucideIcon,
} from "lucide-react";
import { ImageResponse } from "next/og";
import React, { type CSSProperties, type ReactNode } from "react";
import { renderToStaticMarkup } from "react-dom/server";

import layout from "../../../backend/internal/transport/line/navigation/rich_menu.json";

const icons = {
  "rich-menu-clock-in": LogIn,
  "rich-menu-transfer": Repeat2,
  "rich-menu-clock-out": LogOut,
};

const featureIcons: Record<string, LucideIcon> = {
  "missing-punch": CalendarPlus,
  correction: Clock3,
  leave: CalendarDays,
  review: ClipboardCheck,
  repositories: Archive,
};

type Palette = Record<"primary" | "background" | "foreground" | "secondary" | "border", string>;

function iconData(Icon: LucideIcon, size: number, color: string) {
  return `data:image/svg+xml;base64,${Buffer.from(
    renderToStaticMarkup(<Icon size={size} color={color} />),
  ).toString("base64")}`;
}

function MenuCell({
  children,
  palette,
  style,
  active = false,
}: {
  children: ReactNode;
  palette: Palette;
  style: CSSProperties;
  active?: boolean;
}) {
  return (
    <div
      style={{
        display: "flex",
        position: "absolute",
        alignItems: "center",
        justifyContent: "center",
        border: `3px solid ${palette.border}`,
        background: active ? palette.secondary : palette.background,
        color: palette.foreground,
        ...style,
      }}
    >
      {children}
    </div>
  );
}

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
    background: token("background"),
    foreground: token("foreground"),
    secondary: token("secondary"),
    border: token("border"),
  };
  assert.equal(layout.pages.length, 3);
  assert.equal(layout.features.length, 5);
  assert.equal(layout.tabWidths.length, layout.pages.length);
  assert.equal(
    layout.tabWidths.reduce((sum, width) => sum + width, 0),
    layout.width,
  );
  await mkdir(output, { recursive: true });

  const variants = [layout.functionPage, ...layout.pages];
  for (const current of variants) {
    const Icon = icons[current.alias as keyof typeof icons];
    const functions = current.alias === layout.functionPage.alias;
    assert(functions || Icon, `Missing Lucide icon for ${current.alias}`);
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
          <MenuCell
            key={page.alias}
            palette={palette}
            active={current.alias === page.alias}
            style={{
              left: layout.tabWidths.slice(0, index).reduce((sum, width) => sum + width, 0),
              top: 0,
              width: layout.tabWidths[index],
              height: layout.tabHeight,
            }}
          >
            <span style={{ fontSize: 84, fontWeight: current.alias === page.alias ? 700 : 500 }}>
              {page.label}
            </span>
          </MenuCell>
        ))}
        {functions
          ? layout.features.map((feature) => {
              const FeatureIcon = featureIcons[feature.feature];
              assert(FeatureIcon, `Missing Lucide icon for ${feature.feature}`);
              return (
                <MenuCell
                  key={feature.feature}
                  palette={palette}
                  style={{
                    left: feature.bounds.x,
                    top: feature.bounds.y,
                    width: feature.bounds.width,
                    height: feature.bounds.height,
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      alignItems: "center",
                      gap: 28,
                    }}
                  >
                    <img
                      alt=""
                      width={156}
                      height={156}
                      src={iconData(FeatureIcon, 156, palette.primary)}
                    />
                    <span style={{ fontSize: 84, fontWeight: 600 }}>{feature.label}</span>
                  </div>
                </MenuCell>
              );
            })
          : Icon && (
              <MenuCell
                palette={palette}
                style={{
                  left: 0,
                  top: layout.tabHeight,
                  width: layout.width,
                  height: footerY - layout.tabHeight,
                  flexDirection: "column",
                  gap: 36,
                }}
              >
                <img alt="" width={180} height={180} src={iconData(Icon, 180, palette.primary)} />
                <span style={{ fontSize: 104, fontWeight: 700 }}>{current.label}</span>
                <span style={{ fontSize: 44 }}>{current.hint}</span>
              </MenuCell>
            )}
        {[
          { label: layout.functionPage.label, Icon: Grid3X3 },
          { label: layout.miniAppLabel, Icon: Smartphone },
        ].map((item, index) => (
          <MenuCell
            key={item.label}
            palette={palette}
            active={index === 0 && functions}
            style={{
              left: index * (layout.width / 2),
              top: footerY,
              width: layout.width / 2,
              height: layout.footerHeight,
              gap: 36,
            }}
          >
            <img alt="" width={112} height={112} src={iconData(item.Icon, 112, palette.primary)} />
            <span style={{ fontSize: 72, fontWeight: 600 }}>{item.label}</span>
          </MenuCell>
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
