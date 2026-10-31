// Keys belong to a row's lifetime, so deleting/reordering a row preserves its siblings.
let sequence = 0;
const identities = new WeakMap<object, string>();
export function newRowKey(): string { return `row-${++sequence}`; }
export function rowKey(row: object): string {
  let key = identities.get(row);
  if (!key) { key = newRowKey(); identities.set(row, key); }
  return key;
}
