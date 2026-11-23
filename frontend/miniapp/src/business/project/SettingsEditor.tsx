import type { components } from "@repo/api-client";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { FormField } from "@/shared/FormField";
import { NativeSelect, NativeSelectOption } from "@/shared/ui/native-select";

type Settings = components["schemas"]["ProjectSettings"];
const newId = () => crypto.randomUUID().replaceAll("-", "");

export function SettingsEditor({ value, onChange, editable }: {
  value: Settings; onChange: (next: Settings) => void; editable: boolean;
}) {
  const { fields, views } = value;
  function changeField(id: string, change: Partial<Settings["fields"][number]>) {
    onChange({ ...value, fields: fields.map(field => field.id === id ? { ...field, ...change } : field) });
  }
  function changeView(id: string, change: Partial<Settings["views"][number]>) {
    onChange({ ...value, views: views.map(view => view.id === id ? { ...view, ...change } : view) });
  }
  return <fieldset className="grid gap-3" disabled={!editable}>
    <legend>欄位與檢視</legend>
    {fields.map(field => <div key={field.id} className="grid gap-2">
      <FormField>欄位名稱<Input required maxLength={80} value={field.name} onChange={e => changeField(field.id, {name:e.target.value})} /></FormField>
      <FormField>選項（每行一項）<Textarea value={field.options.join("\n")} onChange={e => changeField(field.id, {options:e.target.value ? e.target.value.split("\n") : []})} /></FormField>
      {editable && <Button type="button" variant="outline" disabled={views.some(view => view.fieldId === field.id)}
        onClick={() => onChange({...value, fields:fields.filter(item => item.id !== field.id)})}>移除欄位</Button>}
    </div>)}
    {editable && <Button type="button" variant="outline" disabled={fields.length >= 32} onClick={() => onChange({...value, fields:[...fields, {id:newId(), name:"", options:[]}]})}>新增欄位</Button>}
    {views.map(view => <div key={view.id} className="grid gap-2">
      <FormField>檢視名稱<Input required maxLength={80} value={view.name} onChange={e => changeView(view.id, {name:e.target.value})} /></FormField>
      <FormField>檢視欄位<NativeSelect value={view.fieldId} onChange={e => changeView(view.id, {fieldId:e.target.value})}>
        {fields.map(field => <NativeSelectOption key={field.id} value={field.id}>{field.name}</NativeSelectOption>)}
      </NativeSelect></FormField>
      {editable && <Button type="button" variant="outline" onClick={() => onChange({...value, views:views.filter(item => item.id !== view.id)})}>移除檢視</Button>}
    </div>)}
    {editable && <Button type="button" variant="outline" disabled={!fields.length || views.length >= 32}
      onClick={() => onChange({...value, views:[...views, {id:newId(), name:"", fieldId:fields[0].id}]})}>新增檢視</Button>}
  </fieldset>;
}
