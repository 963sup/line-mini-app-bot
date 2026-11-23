import React, {
  createContext,
  useCallback,
  useContext,
  useLayoutEffect,
  useRef,
  useState,
  useId,
  useEffect,
} from "react";
import { useDiscardChanges } from "@/shared/useDiscardChanges";

type Guard = { dirty: boolean; blocked: boolean; focus: () => void };
type Navigation = { register: (id: string, guard: Guard | null) => void; canWrite: () => boolean };
export const RepositoryNavigationContext = createContext<Navigation | null>(null);

/** Only navigation is shared. Each business retains its writes and recovery rules. */
export function useRepositoryNavigation() {
  const guards = useRef(new Map<string, Guard>());
  const [, refresh] = useState(0);
  const navigating = useRef(false);
  const register = useCallback((id: string, guard: Guard | null) => {
    const previous = guards.current.get(id);
    if (guard) guards.current.set(id, guard);
    else guards.current.delete(id);
    if (previous?.dirty !== guard?.dirty || previous?.blocked !== guard?.blocked)
      refresh((value) => value + 1);
  }, []);
  const dirty = [...guards.current.values()].some((guard) => guard.dirty);
  const blocked = [...guards.current.values()].some((guard) => guard.blocked);
  const discard = useDiscardChanges(dirty, () =>
    [...guards.current.values()].find((guard) => guard.dirty)?.focus(),
  );
  const confirming = useRef(discard.isConfirming);
  confirming.current = discard.isConfirming;
  const canWrite = useCallback(() => !navigating.current && !confirming.current(), []);
  const context = React.useMemo(() => ({ register, canWrite }), [register, canWrite]);
  useEffect(() => {
    if (!dirty && !blocked) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty, blocked]);
  return {
    context,
    blocked,
    dialog: discard.dialog,
    settled: () => {
      navigating.current = false;
    },
    navigate: (next: () => void) => {
      if ([...guards.current.values()].some((guard) => guard.blocked)) return;
      discard.navigate(() => {
        navigating.current = true;
        next();
      });
    },
  };
}

export function useRepositoryGuard(dirty: boolean, focus: () => void) {
  const context = useContext(RepositoryNavigationContext),
    id = useId();
  const current = useRef<Guard>({ dirty, focus, blocked: false });
  current.current.dirty = dirty;
  current.current.focus = focus;
  useLayoutEffect(() => {
    context?.register(id, { ...current.current });
  });
  useLayoutEffect(() => () => context?.register(id, null), [context, id]);
  const block = useCallback(
    (blocked: boolean) => {
      current.current.blocked = blocked;
      context?.register(id, { ...current.current });
    },
    [context, id],
  );
  const canWrite = useCallback(() => context?.canWrite() !== false, [context]);
  return { block, canWrite };
}
