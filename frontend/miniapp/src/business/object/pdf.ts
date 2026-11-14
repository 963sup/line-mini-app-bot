import { untilAbort } from "@/shared/http/lifecycle";

export type PdfPage = { number: number; text: string };
export const PDF_LIMITS = { bytes: 10 * 1024 * 1024, pages: 50, textBytes: 128 * 1024, timeoutMs: 30000 };
type PdfLibrary = typeof import("pdfjs-dist/legacy/build/pdf.mjs");

// Only called from a browser file handler. Importing this module never evaluates PDF.js on the server.
export async function extractPdf(file: File, signal: AbortSignal, load: () => Promise<PdfLibrary> = () => import("pdfjs-dist/legacy/build/pdf.mjs")): Promise<{ pages: PdfPage[]; fileHash: string }> {
  if (file.size > PDF_LIMITS.bytes || !file.size || (file.type && file.type !== "application/pdf")) throw new Error("請選擇10 MiB以內的文字型PDF。");
  const controller = new AbortController(), stop = () => controller.abort();
  signal.addEventListener("abort", stop, { once: true }); if (signal.aborted) stop();
  const timer = setTimeout(stop, PDF_LIMITS.timeoutMs);
  let task: ReturnType<PdfLibrary["getDocument"]> | undefined, worker: InstanceType<PdfLibrary["PDFWorker"]> | undefined;
  const cleanup = () => { void task?.destroy().catch(() => {}); worker?.destroy(); };
  controller.signal.addEventListener("abort", cleanup, { once: true });
  try {
    const bytes = new Uint8Array(await untilAbort(file.arrayBuffer(), controller.signal));
    if (new TextDecoder().decode(bytes.subarray(0, 5)) !== "%PDF-") throw new Error("檔案不是有效PDF。");
    const hash = await untilAbort(crypto.subtle.digest("SHA-256", bytes), controller.signal);
    const fileHash = Array.from(new Uint8Array(hash), value => value.toString(16).padStart(2, "0")).join("");
    const pdfjs = await untilAbort(load(), controller.signal);
    pdfjs.GlobalWorkerOptions.workerSrc = "/assets/pdfjs/pdf.worker.min.mjs";
    worker = new pdfjs.PDFWorker();
    // PDF.js 6.5 has no isEvalSupported option. Only text extraction is used; no rendering or JS-action APIs.
    task = pdfjs.getDocument({ data: bytes, worker, cMapUrl: "/assets/pdfjs/cmaps/", cMapPacked: true, standardFontDataUrl: "/assets/pdfjs/standard_fonts/", wasmUrl: "/assets/pdfjs/wasm/", enableXfa: false, disableFontFace: true, useWorkerFetch: false, stopAtErrors: true });
    const pdf = await untilAbort(task.promise, controller.signal);
    if (pdf.numPages > PDF_LIMITS.pages) throw new Error("PDF最多50頁，請先在本機拆分。");
    const pages: PdfPage[] = []; let total = 0;
    for (let number = 1; number <= pdf.numPages; number++) {
      const page = await untilAbort(pdf.getPage(number), controller.signal);
      try {
        const content = await untilAbort(page.getTextContent(), controller.signal);
        let text = "", lastY: number | undefined;
        for (const item of content.items) {
          if (!("str" in item)) continue;
          const y = item.transform[5];
          if (lastY !== undefined && Math.abs(lastY - y) > 2 && !text.endsWith("\n")) text += "\n";
          text += item.str + (item.hasEOL ? "\n" : " "); lastY = y;
        }
        text = text.trim(); total += new TextEncoder().encode(text).length;
        if (total > PDF_LIMITS.textBytes) throw new Error("PDF文字超過128 KiB，請先在本機拆分。");
        pages.push({ number, text });
      } finally { page.cleanup(); }
    }
    if (!pages.some(page => page.text.trim())) throw new Error("此PDF沒有可擷取文字；掃描影像或加密PDF不支援。");
    return { pages, fileHash };
  } catch (error) {
    if (controller.signal.aborted) throw new Error("PDF擷取已取消或超過30秒。");
    if (error instanceof Error && error.name === "PasswordException") throw new Error("不支援密碼保護PDF。");
    throw error;
  } finally {
    clearTimeout(timer); signal.removeEventListener("abort", stop); controller.signal.removeEventListener("abort", cleanup); cleanup();
  }
}
