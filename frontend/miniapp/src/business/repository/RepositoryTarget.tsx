import { useEffect } from "react";
import { useRepositories } from "./RepositoryContext";
import { FormField } from "@/shared/FormField";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";
import { ReadFeedback } from "@/shared/ReadFeedback";
import { Button } from "@/shared/ui/button";

export function RepositoryTarget({ value, onChange, disabled }: { value: string; onChange: (id: string) => void; disabled: boolean }) {
  const repositories = useRepositories();
  useEffect(() => { repositories.ensure(); }, []);
  return <>
    <FormField>目標儲存庫<NativeSelect required value={value} disabled={disabled || !repositories.items.length} onChange={event => onChange(event.target.value)}>
      <NativeSelectOption value="">請選擇要建立物件的儲存庫</NativeSelectOption>
      {repositories.items.map(item => <NativeSelectOption key={item.id} value={item.id}>{item.name || "群組空間"}</NativeSelectOption>)}
    </NativeSelect></FormField>
    <ReadFeedback busy={repositories.busy} message={repositories.status} label="正在讀取儲存庫……" onRetry={repositories.catalogFailed && !disabled ? repositories.retryCatalog : undefined} />
    {repositories.cursor && <Button type="button" variant="outline" disabled={disabled || repositories.busy} onClick={repositories.loadMore}>載入更多儲存庫</Button>}
  </>;
}
