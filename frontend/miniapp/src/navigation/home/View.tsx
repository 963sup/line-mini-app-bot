import Link from "next/link";
import { ChevronRight, ClipboardList, FolderOpen, Layers, List } from "lucide-react";
import { PagePanel } from "@/shared/PagePanel";
import { Item, ItemContent, ItemMedia, ItemTitle } from "@/shared/ui/item";

const destinations = [
  { title: "動態", href: "/miniapp/feed", icon: List },
  { title: "儲存庫", href: "/miniapp/repositories", icon: FolderOpen },
  { title: "專案", href: "/miniapp/projects", icon: Layers },
  { title: "出勤紀錄", href: "/miniapp/attendance?tab=history", icon: ClipboardList },
];

export function View() {
  return (
    <PagePanel heading="常用功能">
      <nav aria-label="首頁常用功能">
        <ul className="grid gap-2 sm:grid-cols-2">
          {destinations.map(({ title, href, icon: Icon }) => (
            <li key={href}>
              <Item asChild variant="outline" size="sm" className="min-h-11">
                <Link href={href}>
                  <ItemMedia className="text-muted-foreground">
                    <Icon aria-hidden="true" />
                  </ItemMedia>
                  <ItemContent>
                    <ItemTitle>{title}</ItemTitle>
                  </ItemContent>
                  <ChevronRight className="size-4 text-muted-foreground" aria-hidden="true" />
                </Link>
              </Item>
            </li>
          ))}
        </ul>
      </nav>
    </PagePanel>
  );
}
