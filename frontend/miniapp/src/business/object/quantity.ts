// Forms validate unsaved decimal input before sending it to the domain.
const decimal = /^(?:0|[1-9][0-9]{0,17})(?:\.[0-9]{1,6})?$/;

function scaled(value: string): bigint {
  const [whole, fraction = ""] = value.split(".");
  return BigInt(whole + fraction.padEnd(6, "0"));
}

export function isPositiveQuantity(value: string): boolean {
  return decimal.test(value) && scaled(value) > 0n;
}

export function isValidCompletedQuantity(value: string | undefined, total: string): boolean {
  return value !== undefined && decimal.test(value) && decimal.test(total) && scaled(value) <= scaled(total);
}
