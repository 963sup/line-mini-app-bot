// jsdom lacks the browser surfaces used by Radix portals and sidebar sizing.
for (const key of ["DocumentFragment", "Element", "Node", "CustomEvent", "MutationObserver", "HTMLInputElement", "HTMLFormElement", "NodeFilter"] as const) {
  Object.defineProperty(globalThis, key, { value: window[key], configurable: true });
}
Object.assign(globalThis, {
  getComputedStyle: window.getComputedStyle.bind(window),
  requestAnimationFrame: (callback: FrameRequestCallback) => setTimeout(() => callback(0), 0),
  cancelAnimationFrame: clearTimeout,
  ResizeObserver: class { observe() {} unobserve() {} disconnect() {} },
});
window.matchMedia = (query) => ({ matches: false, media: query, onchange: null,
  addListener() {}, removeListener() {}, addEventListener() {}, removeEventListener() {}, dispatchEvent() { return false; } });

// Radix tabs activate on primary mousedown. HTMLElement.click() alone omits it.
export function pressButton(button: HTMLButtonElement) {
  if (button.disabled) return;
  button.dispatchEvent(new window.MouseEvent("mousedown", { bubbles: true, button: 0 }));
  button.click();
}
