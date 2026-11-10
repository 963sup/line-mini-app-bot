import { JSDOM } from "jsdom";
const dom = new JSDOM("<!doctype html><html><body></body></html>", {
  url: "https://example.test/miniapp",
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
] as const)
  Object.defineProperty(globalThis, key, {
    value: dom.window[key],
    configurable: true,
    writable: true,
  });
Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
Object.defineProperty(window, "isSecureContext", { value: true });
