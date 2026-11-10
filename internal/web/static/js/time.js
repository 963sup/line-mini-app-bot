export const taiwan = value => value && !value.startsWith("0001-")
  ? new Date(new Date(value).getTime() + 28800000).toISOString().slice(0, 19) : "";
export const displayTime = value => taiwan(value).replace("T", " ") || "尚未下班";
