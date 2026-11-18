export function taiwan(value: string) {
  if (!value || value.startsWith("0001")) return "";
  const date = new Date(value);
  return Number.isFinite(date.getTime())
    ? new Date(date.getTime() + 28800000).toISOString().slice(0, 19)
    : "";
}
export function displayTime(value: string) {
  return taiwan(value).replace("T", " ") || "尚未下班";
}
export function timestamp(value: string) {
  if (!value) throw new Error("請填寫完整的時間。");
  const date = new Date(`${value.length === 16 ? value + ":00" : value}+08:00`);
  if (!Number.isFinite(date.getTime())) throw new Error("時間格式無效。");
  return date.toISOString();
}
