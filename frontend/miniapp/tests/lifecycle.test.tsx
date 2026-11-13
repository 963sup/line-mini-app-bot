import { test } from "node:test";
import assert from "node:assert/strict";
import { untilAbort } from "../src/shared/http/lifecycle";

function deferred() {
  let resolve!: (value: string) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<string>((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}

function trackedSignal() {
  const controller = new AbortController();
  const signal = controller.signal;
  let added = 0, removed = 0;
  const add = signal.addEventListener.bind(signal);
  const remove = signal.removeEventListener.bind(signal);
  signal.addEventListener = ((...args: Parameters<typeof add>) => {
    added++; return add(...args);
  }) as typeof add;
  signal.removeEventListener = ((...args: Parameters<typeof remove>) => {
    removed++; return remove(...args);
  }) as typeof remove;
  return { controller, signal, counts: () => ({ added, removed }) };
}

const flushRejections = () => new Promise<void>((resolve) => setImmediate(resolve));

test("pre-aborted work rejects immediately and still observes its original promise", async (t) => {
  for (const result of ["already-rejected", "late-rejected", "late-resolved"]) {
    await t.test(result, async () => {
      const { controller, signal, counts } = trackedSignal();
      controller.abort();
      const work = deferred();
      if (result === "already-rejected") work.reject(new Error("transport failure"));
      await assert.rejects(untilAbort(work.promise, signal), { name: "AbortError" });
      if (result === "late-rejected") work.reject(new Error("transport failure"));
      if (result === "late-resolved") work.resolve("late success");
      // node:test reports an unhandled rejection as a test failure. Do not
      // consume the original promise here: observing it is untilAbort's job.
      await flushRejections();
      assert.equal(counts().added, 0);
    });
  }
});

test("cancellation rejects before late transport failure and releases its listener", async () => {
  const { controller, signal, counts } = trackedSignal();
  const work = deferred();
  const aborted = assert.rejects(untilAbort(work.promise, signal), { name: "AbortError" });
  controller.abort();
  await aborted;
  work.reject(new Error("late transport failure"));
  await flushRejections();
  assert.deepEqual(counts(), { added: 1, removed: 1 });
});

test("uncancelled work preserves values and the original error, releasing listeners", async () => {
  const success = trackedSignal();
  assert.equal(await untilAbort(Promise.resolve("value"), success.signal), "value");
  assert.deepEqual(success.counts(), { added: 1, removed: 1 });
  const failure = trackedSignal();
  const error = new Error("transport failure");
  await assert.rejects(untilAbort(Promise.reject(error), failure.signal), (actual) => actual === error);
  assert.deepEqual(failure.counts(), { added: 1, removed: 1 });
});
