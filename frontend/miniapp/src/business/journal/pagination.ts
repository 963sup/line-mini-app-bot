export async function collect<T>(fetch: (cursor: string) => Promise<{items:T[];nextCursor:string}>): Promise<T[]> {
  const all:T[]=[]; let cursor="";
  do { const page=await fetch(cursor); all.push(...page.items); cursor=page.nextCursor; } while(cursor);
  return all;
}

