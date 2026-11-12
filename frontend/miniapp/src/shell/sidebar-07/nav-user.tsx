"use client";

import { ChevronsUpDown, UserRound } from "lucide-react";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/shared/ui/sidebar";

export type AccountMenu = {
  name: string;
  detail?: string;
  status: string;
  actions: {
    id?: string;
    label: string;
    onSelect?: () => void;
    url?: string;
    disabled?: boolean;
  }[];
};
export function NavUser({ user }: { user: AccountMenu }) {
  const { isMobile } = useSidebar();
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              aria-label="身分選單"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar className="size-8 rounded-lg">
                <AvatarFallback className="rounded-lg">
                  <UserRound className="size-4" />
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{user.name}</span>
                <span
                  id="mini-status"
                  role="status"
                  aria-live="polite"
                  className="truncate text-xs"
                >
                  {user.status}
                </span>
              </div>
              <ChevronsUpDown className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 max-w-[calc(100vw-2rem)] rounded-lg"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuGroup>
              <DropdownMenuLabel className="font-normal">
                <div className="grid gap-1">
                  <span className="font-medium">{user.name}</span>
                  <span className="break-all text-xs">{user.detail}</span>
                  <span className="text-xs text-muted-foreground">
                    {user.status}
                  </span>
                </div>
              </DropdownMenuLabel>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              {user.actions.map((action) => (
                <DropdownMenuItem
                  key={action.label}
                  id={action.id}
                  disabled={action.disabled}
                  onSelect={action.onSelect}
                  asChild={!!action.url}
                >
                  {action.url ? (
                    <a href={action.url}>{action.label}</a>
                  ) : (
                    action.label
                  )}
                </DropdownMenuItem>
              ))}
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
