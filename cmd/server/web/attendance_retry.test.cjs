const fs = require("fs");
const path = require("path");
const vm = require("vm");
const assert = require("assert");
const nodes = Object.fromEntries(["punch", "punch-status", "attendance", "logout"].map(id => [id, {
  disabled: false, hidden: false, textContent: "", handlers: {},
  addEventListener(name, fn) { this.handlers[name] = fn; },
  setAttribute() {}, removeAttribute() {},
}]));
const events = {};
const callbacks = [];
const requests = [];
const timers = new Map();
let nextTimer = 0;
const context = {
  document: { querySelector: selector => nodes[selector.slice(1)] },
  window: { location: { search: `?request=${"a".repeat(64)}`, origin: "https://example.test" },
    isSecureContext: true, liff: { isLoggedIn: () => true },
    addEventListener(name, fn) { events[name] = fn; } },
  navigator: { geolocation: { getCurrentPosition(ok, error, options) { callbacks.push({ ok, error, options }); } } },
  liff: { getIDToken: () => "token", isInClient: () => false },
  URLSearchParams, URL, AbortController, Date, Number, Math, JSON,
  setTimeout(fn) { timers.set(++nextTimer, fn); return nextTimer; },
  clearTimeout(id) { timers.delete(id); },
  fetch(url, options) { return new Promise(resolve => requests.push({ url, options, resolve })); },
};
vm.runInNewContext(fs.readFileSync(path.join(__dirname, "attendance.js"), "utf8"), context);
const flush = () => new Promise(setImmediate);
const position = { coords: { latitude: 25, longitude: 121, accuracy: 5 }, timestamp: Date.now() };
(async () => {
  // The old public request URL must not activate attendance.
  context.window.location.search = `?request=${"a".repeat(64)}`;
  events["miniapp:identity-ready"]();
  assert.equal(requests.length, 0);
  assert.equal(nodes.attendance.hidden, true);
  context.window.location.search = `?launch=${"b".repeat(64)}`;
  nodes.punch.handlers.click(); nodes.punch.handlers.click();
  await flush();
  assert.equal(requests.length, 1);
  assert.match(requests[0].url, /\/attendance\/launch$/);
  requests[0].resolve({ ok: true, json: async () => ({ requestId: "a".repeat(64) }) });
  await flush();
  assert.equal(callbacks.length, 1);
  assert.equal(callbacks[0].options.timeout, 10000);
  for (const fn of [...timers.values()]) fn();
  assert.equal(nodes.punch.disabled, false);
  callbacks[0].ok(position);
  await flush();
  assert.equal(requests.length, 1);
  nodes.punch.handlers.click();
  await flush();
  assert.equal(callbacks.length, 2);
  assert.equal(requests.length, 1); // Stable exchange result reused.
  callbacks[1].ok(position);
  await flush();
  assert.equal(requests.length, 2);
  assert.equal(JSON.parse(requests[1].options.body).requestId, "a".repeat(64));
  for (const fn of [...timers.values()]) fn();
  assert.equal(nodes.punch.disabled, false);
  assert.match(nodes["punch-status"].textContent, /尚未確認/);
  requests[1].resolve({ ok: true, json: async () => ({ code: "clocked_in", message: "success" }) });
  await flush();
  assert.match(nodes["punch-status"].textContent, /尚未確認/);
  nodes.punch.handlers.click();
  await flush();
  assert.equal(callbacks.length, 3);
  events.pagehide();
  assert.equal(nodes.punch.disabled, false);
  assert.match(nodes["punch-status"].textContent, /頁面切換/);
  callbacks[2].ok(position);
  await flush();
  assert.equal(requests.length, 2);
  nodes.punch.handlers.click();
  await flush();
  assert.equal(requests.length, 3);
  assert.match(requests[2].url, /\/attendance\/launch$/);
  requests[2].resolve({ ok: true, json: async () => ({ requestId: "a".repeat(64) }) });
  await flush();
  callbacks[3].ok(position);
  await flush();
  assert.equal(requests.length, 4);
  events.pagehide();
  assert.equal(nodes.punch.disabled, false);
  assert.match(nodes["punch-status"].textContent, /尚未確認/);
  assert.equal(requests[3].options.signal.aborted, true);
  requests[3].resolve({ ok: true, json: async () => ({ code: "clocked_in", message: "success" }) });
  await flush();
  assert.match(nodes["punch-status"].textContent, /尚未確認/);
  console.log("PASS launch retry, old URL rejected, double click, timeout, pagehide, late callbacks, same request ID");
})().catch(error => { console.error(error); process.exitCode = 1; });
