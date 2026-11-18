import type { PdfPage } from "./pdf";
export type PositionedText = { str: string; x: number; y: number; width: number };
// This supported order layout has explicit column headers. Unknown layouts fail closed.
export function orderTable(number: number, items: PositionedText[]): PdfPage | null {
  const header = (name: string) => items.find(item => item.str.trim() === name);
  const item = header("項次"), name = header("料號/品名"), quantity = header("數量"), unit = header("單位"), price = header("單價");
  if (!item && !name && !quantity && !unit && !price) return null;
  if (!item || !name || !quantity || !unit || !price || !(item.x < name.x && name.x < quantity.x && quantity.x < unit.x && unit.x < price.x) || [name, quantity, unit, price].some(value => Math.abs(value.y - item.y) > 3)) throw new Error("此PDF明細欄位無法安全辨識，目前不支援此版型。");
  const right = (unit.x + unit.width + price.x) / 2;
  const end = items.filter(value => /^(未稅總計|含稅總價|採購安全衛生審核須知|Page)$/.test(value.str.trim()) && value.y > item.y).reduce((y, value) => Math.min(y, value.y), Infinity);
  if (!Number.isFinite(end)) throw new Error("此PDF缺少可辨識的明細結束位置，目前不支援此版型。");
  const selected = items.filter(value => value.y > item.y + 5 && value.y < end - 3 && value.x >= item.x - 2 && value.x + value.width < right && value.str.trim());
  const lines: { y: number; items: PositionedText[] }[] = [];
  for (const value of selected.sort((a, b) => a.y - b.y || a.x - b.x)) {
    const line = lines.find(line => Math.abs(line.y - value.y) < 3);
    if (line) line.items.push(value); else lines.push({ y: value.y, items: [value] });
  }
  const text = lines.map(line => line.items.sort((a, b) => a.x - b.x).map(value => value.str).join(" ").trim()).filter(value => !/^\((?:Cost Ref|Customer PO):/i.test(value)).join("\n");
  if (!text) throw new Error("此PDF沒有可辨識的工程明細。");
  return { number, text };
}
export function orderRevision(items: PositionedText[]): string {
  const label = items.filter(item => item.str.trim() === "版本");
  if (label.length !== 1) return "";
  const value = items.filter(item => item.x > label[0].x + label[0].width && Math.abs(item.y - label[0].y) < 3).sort((a, b) => a.x - b.x)[0];
  return value && value.x - label[0].x - label[0].width < 45 && /^[A-Za-z0-9._-]{1,20}$/.test(value.str.trim()) ? value.str.trim() : "";
}
