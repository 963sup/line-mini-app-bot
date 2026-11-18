import test from "node:test";
import assert from "node:assert/strict";
import { orderTable, orderRevision, type PositionedText } from "../src/business/object/pdf-table";
const item = (str: string, x: number, y: number, width = 15): PositionedText => ({str,x,y,width});
const headers = [item("項次",10,100),item("料號/品名",50,100),item("數量",250,100),item("單位",300,100),item("單價",380,100)];
test("only supported table detail columns survive and continuation preserves page and source unit", () => {
  const detail = [item("10",10,120),item("測試物件",50,120),item("2",250,120),item("M2",300,120),item("9999",380,120),item("聯絡資料",10,80),item("未稅總計",10,160),item("外部條款",10,180)];
  assert.deepEqual(orderTable(2,[...headers,...detail]), {number:2,text:"10 測試物件 2 M2"});
  assert.equal(orderTable(1,[item("普通文字",10,10)]),null);
  assert.throws(() => orderTable(1,headers), /結束位置/);
  assert.throws(() => orderTable(1,headers.filter(value => value.str !== "單位")), /版型/);
});
test("revision comes only from an unambiguous adjacent version label including zero", () => {
 assert.equal(orderRevision([item("版本",10,10), item("0",30,10)]),"0");
 assert.equal(orderRevision([item("版本",10,10), item("條款",300,10)]),"");
 assert.equal(orderRevision([item("v1",30,10)]),"");
});
