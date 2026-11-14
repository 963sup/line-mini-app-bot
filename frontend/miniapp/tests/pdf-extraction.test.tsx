import test from "node:test";
import assert from "node:assert/strict";
import { extractPdf, PDF_LIMITS } from "../src/business/object/pdf";
const pdf = () => new File(["%PDF-1.7\nlocal-test"], "sample.pdf", { type: "application/pdf" });
function library(options: { pageCount?: number; text?: string; stall?: boolean; password?: boolean } = {}) {
  let destroyed = 0, terminated = 0, cleaned = 0; const requests: Record<string, unknown>[] = [];
  const fake = {
    GlobalWorkerOptions: { workerSrc: "" },
    PDFWorker: class { destroy() { terminated++; } },
    getDocument(input: Record<string, unknown>) {
      requests.push(input);
      return { destroy: async () => { destroyed++; }, promise: options.password ? Promise.reject(Object.assign(new Error("private"), { name: "PasswordException" })) : options.stall ? new Promise(() => {}) : Promise.resolve({ numPages: options.pageCount || 1, getPage: async () => ({ cleanup() { cleaned++; }, getTextContent: async () => ({ items: [{ str: options.text ?? "first", transform: [1, 0, 0, 1, 0, 20], hasEOL: true }, { str: "second", transform: [1, 0, 0, 1, 0, 10], hasEOL: false }] }) }) }) };
    },
  };
  return { fake, load: async () => fake as never, requests, counts: () => ({ destroyed, terminated, cleaned }) };
}
test("PDF extraction uses binary local data, fixed same-origin assets, passive settings, and preserves page/line boundaries", async () => {
  const lib = library(); const result = await extractPdf(pdf(), new AbortController().signal, lib.load);
  assert.equal(result.pages[0].text, "first\nsecond"); assert.match(result.fileHash, /^[a-f0-9]{64}$/);
  const options = lib.requests[0]; assert.ok(options.data instanceof Uint8Array); assert.ok(!("url" in options));
  assert.equal(options.enableXfa, false); assert.equal(options.useWorkerFetch, false); assert.equal(options.disableFontFace, true);
  assert.equal(lib.fake.GlobalWorkerOptions.workerSrc, "/assets/pdfjs/pdf.worker.min.mjs");
  assert.equal(options.cMapUrl, "/assets/pdfjs/cmaps/"); assert.equal(options.standardFontDataUrl, "/assets/pdfjs/standard_fonts/");
  assert.deepEqual(lib.counts(), { destroyed: 1, terminated: 1, cleaned: 1 });
});
test("PDF safety rejects bad MIME, magic, size, page and text bounds and unsupported password/empty documents", async () => {
  const signal = new AbortController().signal;
  for (const file of [new File(["%PDF-"], "fake", { type: "text/html" }), new File(["not pdf"], "fake", { type: "application/pdf" }), new File([new Uint8Array(PDF_LIMITS.bytes + 1)], "large", { type: "application/pdf" })]) await assert.rejects(extractPdf(file, signal, async () => { throw new Error("must not load library"); }), /PDF/);
  const tooMany = library({ pageCount: 51 }); await assert.rejects(extractPdf(pdf(), signal, tooMany.load), /50頁/); assert.equal(tooMany.counts().terminated, 1);
  const long = library({ text: "a".repeat(PDF_LIMITS.textBytes + 1) }); await assert.rejects(extractPdf(pdf(), signal, long.load), /128 KiB/); assert.equal(long.counts().cleaned, 1);
  const password = library({ password: true }); await assert.rejects(extractPdf(pdf(), signal, password.load), /密碼/); assert.equal(password.counts().terminated, 1);
  const empty = library({ text: "" }); empty.fake.getDocument = () => ({ destroy: async () => {}, promise: Promise.resolve({ numPages: 1, getPage: async () => ({ cleanup() {}, getTextContent: async () => ({ items: [] }) }) }) });
  await assert.rejects(extractPdf(pdf(), signal, empty.load), /沒有可擷取文字/);
});
test("aborting a blocked PDF load destroys its loading task and worker", async () => {
  const lib = library({ stall: true }), controller = new AbortController();
  const pending = extractPdf(pdf(), controller.signal, lib.load);
  for (let i = 0; i < 20 && !lib.requests.length; i++) await new Promise<void>(resolve => setImmediate(resolve));
  assert.equal(lib.requests.length, 1); controller.abort(); await assert.rejects(pending, /取消/);
  assert.ok(lib.counts().destroyed >= 1); assert.ok(lib.counts().terminated >= 1);
});
