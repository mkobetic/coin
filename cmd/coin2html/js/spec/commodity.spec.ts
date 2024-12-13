import { Amount, Commodity } from "../src/commodity";

test("create commodity", () =>
  expect(new Commodity("CAD", "Canadian Dollar", 2, "")).toBeTruthy());

describe("amount", () => {
  const CAD = new Commodity("CAD", "Canadian Dollar", 2, "");
  test.each([
    [0, "0.00 CAD"],
    [1, "0.01 CAD"],
    [-1, "-0.01 CAD"],
    [200, "2.00 CAD"],
    [-50, "-0.50 CAD"],
    [123456789, "1,234,567.89 CAD"],
    [-12345678, "-123,456.78 CAD"],
  ])(`%#: %i`, (i, expected) => {
    const amt = new Amount(i, CAD);
    expect(amt.toString()).toBe(expected);
  });
});
