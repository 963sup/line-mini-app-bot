"""Render or check the attendance Rich Menu JPGs published by LINE."""
import argparse
import hashlib
from io import BytesIO
import json
from pathlib import Path
import re
from tempfile import TemporaryDirectory

from PIL import Image, ImageDraw, ImageFont, __version__ as pillow_version

ROOT = Path(__file__).resolve().parents[2]
SPEC = ROOT / "backend/internal/transport/line/navigation/attendance-rich-menu.json"
ASSETS = ROOT / "assets/line/richmenu"
FONT = Path("/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc")
FONT_HASH = "b76b0433203017ca80401b2ee0dd69350349871c4b19d504c34dbdd80541690a"


def render(layout, active, font_path):
    page = layout["pages"][active]
    width, height, tab_height = (layout[key] for key in ("width", "height", "tabHeight"))
    image = Image.new("RGB", (width, height), "#F4F6F9")
    draw = ImageDraw.Draw(image)

    def text(x, y, label, size, color, stroke=0):
        font = ImageFont.truetype(font_path, size, index=3)
        draw.text((x, y), label, font=font, fill=color, anchor="mm", stroke_width=stroke)

    x = 0
    for i, tab_width in enumerate(layout["tabWidths"]):
        right = x + tab_width
        if i == active:
            draw.rectangle((x, 0, right - 1, tab_height - 1), fill="#E5EBF1")
            draw.rectangle((x, tab_height - 24, right - 1, tab_height - 1), fill=page["color"])
        elif i:
            draw.rectangle((x, tab_height * 104 // 843, x + 3, tab_height * 739 // 843), fill="#D6DEE7")
        center = x + tab_width // 2
        text(center, tab_height * 350 // 843, layout["pages"][i]["label"], 104, "#172F44")
        text(center, tab_height * 533 // 843, "目前頁面" if i == active else "切換頁面", 56, "#3A5168")
        x = right

    draw.rectangle((0, tab_height, width - 1, height - 1), fill=page["color"])
    remaining = height - tab_height
    text(width // 2, tab_height + remaining * 167 // 843, "出勤操作", 54, "white")
    text(width // 2, tab_height + remaining * 415 // 843, page["label"], 164, "white", 2)
    text(width // 2, tab_height + remaining * 669 // 843, page["hint"], 64, "white")
    output = BytesIO()
    image.save(output, "JPEG", quality=93, optimize=True)
    data = output.getvalue()
    if len(data) > 1024 * 1024:
        raise ValueError(f"{page['alias']} exceeds 1 MiB")
    return data


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("generate", "check"))
    parser.add_argument("--font", type=Path, default=FONT)
    parser.add_argument("--output-dir", type=Path, default=ASSETS)
    args = parser.parse_args()

    if pillow_version != "13.0.0":
        parser.error("Pillow 13.0.0 is required to reproduce the committed JPGs")
    if not args.font.is_file() or hashlib.sha256(args.font.read_bytes()).hexdigest() != FONT_HASH:
        parser.error("use the documented Noto Sans CJK font revision (--font)")

    layout = json.loads(SPEC.read_text(encoding="utf-8"))
    pages, widths = layout["pages"], layout["tabWidths"]
    aliases = [page["alias"] for page in pages]
    if (layout["width"], layout["height"]) != (2500, 1686) or not 0 < layout["tabHeight"] < layout["height"]:
        parser.error("invalid canvas or tab height")
    if len(pages) != 3 or len(widths) != len(pages) or min(widths) <= 0 or sum(widths) != layout["width"]:
        parser.error("invalid tab layout")
    if len(set(aliases)) != len(aliases) or any(not re.fullmatch(r"[a-z0-9][a-z0-9-]{0,31}", alias) for alias in aliases):
        parser.error("invalid or duplicate alias")
    images = {f"{alias}.JPG": render(layout, i, args.font) for i, alias in enumerate(aliases)}

    if args.mode == "check":
        stale = [name for name, data in images.items()
                 if not (path := args.output_dir / name).is_file() or path.read_bytes() != data]
        if stale:
            parser.exit(1, "Missing or stale JPGs: " + ", ".join(stale) + "\nRun rich-menu.py generate.\n")
    else:
        args.output_dir.mkdir(parents=True, exist_ok=True)
        with TemporaryDirectory(dir=args.output_dir) as staging:
            for name, data in images.items():
                (Path(staging) / name).write_bytes(data)
            for name in images:
                (Path(staging) / name).replace(args.output_dir / name)
    print(f"{args.mode}: {len(images)} Rich Menu JPGs")


if __name__ == "__main__":
    main()
