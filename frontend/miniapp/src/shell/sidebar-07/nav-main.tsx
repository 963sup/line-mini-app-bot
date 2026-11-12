"use client";

import { useState } from "react";
import { ChevronRight, type LucideIcon } from "lucide-react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/shared/ui/collapsible";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  useSidebar,
} from "@/shared/ui/sidebar";

export type NavItem = {
  title: string;
  url?: string;
  icon?: LucideIcon;
  isActive?: boolean;
  onSelect?: () => void;
  items?: Omit<NavItem, "items">[];
};

export function NavMain({ items }: { items: NavItem[] }) {
  const { state, isMobile, setOpen, setOpenMobile } = useSidebar();
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({});
  function select(item: NavItem) {
    item.onSelect?.();
    setOpenMobile(false);
  }
  return (
    <nav aria-label="服務導覽">
      <SidebarGroup>
        <SidebarGroupLabel>服務</SidebarGroupLabel>
        <SidebarMenu>
          {items.map((item) =>
            item.items?.length ? (
              <Collapsible
                key={item.title}
                asChild
                open={openGroups[item.title] ?? !!item.isActive}
                onOpenChange={(open) =>
                  setOpenGroups((previous) => ({
                    ...previous,
                    [item.title]: open,
                  }))
                }
                className="group/collapsible"
              >
                <SidebarMenuItem>
                  <CollapsibleTrigger asChild>
                    <SidebarMenuButton
                      tooltip={item.title}
                      isActive={item.isActive}
                      onClick={(event) => {
                        if (!isMobile && state === "collapsed") {
                          event.preventDefault();
                          setOpen(true);
                          setOpenGroups((previous) => ({
                            ...previous,
                            [item.title]: true,
                          }));
                        }
                      }}
                    >
                      {item.icon && <item.icon aria-hidden="true" />}
                      <span>{item.title}</span>
                      <ChevronRight
                        className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"
                        aria-hidden="true"
                      />
                    </SidebarMenuButton>
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <SidebarMenuSub>
                      {item.items.map((subItem) => (
                        <SidebarMenuSubItem key={subItem.title}>
                          <SidebarMenuSubButton
                            asChild
                            isActive={subItem.isActive}
                            aria-current={subItem.isActive ? "page" : undefined}
                            onClick={() => select(subItem)}
                          >
                            {subItem.url ? (
                              <a href={subItem.url}>
                                <span>{subItem.title}</span>
                              </a>
                            ) : (
                              <button type="button">{subItem.title}</button>
                            )}
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      ))}
                    </SidebarMenuSub>
                  </CollapsibleContent>
                </SidebarMenuItem>
              </Collapsible>
            ) : (
              <SidebarMenuItem key={item.title}>
                <SidebarMenuButton
                  asChild
                  isActive={item.isActive}
                  tooltip={item.title}
                  onClick={() => select(item)}
                >
                  <a
                    href={item.url}
                    aria-current={item.isActive ? "page" : undefined}
                  >
                    {item.icon && <item.icon aria-hidden="true" />}
                    <span>{item.title}</span>
                  </a>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ),
          )}
        </SidebarMenu>
      </SidebarGroup>
    </nav>
  );
}
