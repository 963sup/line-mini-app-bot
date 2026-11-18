import type { components } from "@repo/api-client";

export type Journal = components["schemas"]["Journal"];
export type Draft = components["schemas"]["JournalCreation"];
export type Reference = components["schemas"]["JournalReference"];
export const roles = { management: "管理職", control: "管制人員", workers: "施作人員" } as const;
export const workDate = () => new Intl.DateTimeFormat("sv-SE", { timeZone: "Asia/Taipei" }).format(new Date());
