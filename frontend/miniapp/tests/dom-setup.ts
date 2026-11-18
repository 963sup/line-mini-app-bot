import { JSDOM } from "jsdom";
const dom = new JSDOM("<!doctype html><html><body></body></html>", {
  url: "https://example.test/miniapp",
  pretendToBeVisual: true,
});
for (const key of [
  "window",
  "document",
  "navigator",
  "location",
  "HTMLElement",
  "Event",
  "MouseEvent",
  "PageTransitionEvent",
  "getComputedStyle",
] as const)
  Object.defineProperty(globalThis, key, {
    value: dom.window[key],
    configurable: true,
    writable: true,
  });
Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
Object.defineProperty(window, "isSecureContext", { value: true });

// JSDOM has no layout engine. Radix Checkbox observes its hidden form input size.
class TestResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
Object.defineProperty(globalThis, "ResizeObserver", { value: TestResizeObserver, configurable: true, writable: true });
Object.defineProperty(window, "ResizeObserver", { value: TestResizeObserver, configurable: true, writable: true });
