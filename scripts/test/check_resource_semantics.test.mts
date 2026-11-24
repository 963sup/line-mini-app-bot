import assert from "node:assert/strict";
import test from "node:test";
import {
  checkResourceSemantics,
  type ResourceEntry,
} from "../contract/check_resource_semantics.mts";

const resource: ResourceEntry = {
  entity: "PunchResponse",
  collection: "punches",
  path: "/punches",
  owner: "attendance.punch",
  kind: "command-result",
};
const owners = { "attendance.punch": {}, other: {} };
const operations = [{ path: "/punches", owner: "attendance.punch" }];

test("a command result keeps its existing public type and route", () => {
  assert.doesNotThrow(() => checkResourceSemantics([resource], owners, operations));
});

test("resource ownership must match its actual operation", () => {
  assert.throws(
    () => checkResourceSemantics([{ ...resource, owner: "other" }], owners, operations),
    /differs/,
  );
  assert.throws(
    () => checkResourceSemantics([{ ...resource, owner: "missing" }], owners, operations),
    /Unknown resource owner/,
  );
});

test("unknown representations and duplicate routes fail closed", () => {
  assert.throws(
    () =>
      checkResourceSemantics(
        [{ ...resource, kind: "aggregate" as ResourceEntry["kind"] }],
        owners,
        operations,
      ),
    /Unknown resource representation/,
  );
  assert.throws(
    () => checkResourceSemantics([resource, resource], owners, operations),
    /Duplicate/,
  );
});
