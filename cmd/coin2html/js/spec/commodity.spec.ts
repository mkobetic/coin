import {
  Amount,
  Commodities,
  commodity,
  Commodity,
  composeConversions,
  newConversion,
  Price,
} from "../src/commodity";

for (const [id, decimals] of Object.entries({
  USD: 2,
  CAD: 2,
  EUR: 2,
  CZK: 2,
}))
  if (!Commodities[id]) Commodities[id] = new Commodity(id, id, decimals, "");

describe("amount", () => {
  const CAD = commodity`CAD`;
  test.each([
    [0, "0.00 CAD"],
    [1, "0.01 CAD"],
    [-1, "-0.01 CAD"],
    [200, "2.00 CAD"],
    [-50, "-0.50 CAD"],
    [123456789, "1,234,567.89 CAD"],
    [-12345678, "-123,456.78 CAD"],
  ])(`%#: toString %i`, (i, expected) => {
    const amt = new Amount(i, CAD);
    expect(amt.toString()).toBe(expected);
  });

  test.each([
    ["25.00 CAD", "25.00 CAD"],
    ["25.0 CAD", "25.00 CAD"],
    ["25 CAD", "25.00 CAD"],
    ["-25 CAD", "-25.00 CAD"],
    ["0.1 CAD", "0.10 CAD"],
    ["-0.02834 CAD", "-0.02 CAD"],
    ["25.0330 CAD", "25.03 CAD"],
    ["25,033.5 CAD", "25,033.50 CAD"],
  ])(`%#: parse %s`, (input, expected) => {
    expect(Amount.parse(input).toString()).toBe(expected);
  });

  test.each([
    ["4 CAD", 4, 2500],
    ["0.25 CAD", 4, 40000],
    ["4 CAD", 2, 25],
    ["0.25 CAD", 2, 400],
  ])(`%#: reciprocal %s/%d`, (input, decimals, expected) => {
    expect(Amount.parse(input).reciprocal(decimals)).toBe(expected);
  });
});

describe("price", () => {
  test.each([
    ["CAD: 0.75 USD @ 2000-01-01"],
    ["USD: 1.33 CAD @ 2000-01-01"],
    ["USD: 0.99 EUR @ 2010-01-01"],
  ])("toString %s", (input) => {
    const price = Price.parse(input);
    expect(price.toString()).toBe(input);
  });
  test.each([
    ["CAD: 0.75 USD", "USD: 1.33 CAD"],
    ["USD: 1.33 CAD", "CAD: 0.75 USD"],
  ])("reverse %s", (input, expected) => {
    const price = Price.parse(input);
    const reversed = price.reverse().toString().split("@")[0].trim();
    expect(reversed).toBe(expected);
  });
});

describe("conversions", () => {
  test("composition", () => {
    const day = new Date("2000-01-01");
    const dayString = day.toISOString().split("T")[0];
    const cu = newConversion([Price.parse(`CAD: 0.75 USD @ ${dayString}`)]);
    const ue = newConversion([Price.parse(`USD: 0.90 EUR @ ${dayString}`)]);
    const ce = composeConversions(cu, ue);
    expect(ce.direction).toBe("CAD => USD => EUR");
    expect(ce(day).toString()).toBe("CAD: 0.68 EUR @ 2000-01-01");
    expect(() => composeConversions(ue, cu)).toThrow();

    const ez = newConversion([Price.parse(`EUR: 25.00 CZK @ ${dayString}`)]);
    const cz = composeConversions(cu, composeConversions(ue, ez));
    expect(cz.direction).toBe("CAD => USD => EUR => CZK");
    expect(cz(day).toString()).toBe("CAD: 16.88 CZK @ 2000-01-01");
  });

  describe("amount conversions", () => {
    const CAD = new Commodity("CAD", "CAD", 2, "");
    const USD = new Commodity("USD", "USD", 2, "");
    const EUR = new Commodity("EUR", "EUR", 2, "");
    const CZK = new Commodity("CZK", "CZK", 2, "");
    const day = new Date("2000-01-01");
    for (const [com, val, com2] of [
      [CAD, 75, USD],
      [USD, 90, EUR],
      [EUR, 2500, CZK],
      [CAD, 68, EUR],
    ] as [Commodity, number, Commodity][]) {
      const p = new Price(com, day, new Amount(val, com2), "");
      com.prices.push(p);
      com2.prices.push(p.reverse());
    }
    test.each([
      [350, CAD, 350, USD, "8.16 CAD"],
      [350, USD, 350, CAD, "6.13 USD"],
      [350, EUR, 350, USD, "6.65 EUR"],
      [350, USD, 350, EUR, "7.39 USD"],
      [350, EUR, 3500, CZK, "4.90 EUR"],
      [3500, CZK, 350, EUR, "122.50 CZK"],
      [350, CAD, 350, EUR, "8.65 CAD"],
      [350, EUR, 350, CAD, "5.88 EUR"],
      [3500, CZK, 350, CAD, "94.50 CZK"],
      [3500, CZK, 350, USD, "113.75 CZK"],
      [350, CAD, 3500, CZK, "5.60 CAD"],
      [350, USD, 3500, CZK, "4.90 USD"],
    ])(`%#: %d %s + %d %s`, (fv, fc, tv, tc, exp) => {
      expect(new Amount(fv, fc).addIn(new Amount(tv, tc), day).toString()).toBe(
        exp
      );
    });
    test.each([
      [CAD, CZK, "17.00 CZK"],
      [USD, CZK, "22.50 CZK"],
      [CZK, USD, "0.04 USD"],
      [CZK, CAD, "0.06 CAD"],
    ])(`%#: %s -> %s`, (from, to, exp) => {
      expect(to.convert(new Amount(100, from), day).toString()).toBe(exp);
    });
    test.each([
      [CAD, ["CAD => USD", "CAD => EUR", "CAD => EUR => CZK"]],
      [USD, ["USD => CAD", "USD => EUR", "USD => EUR => CZK"]],
      [EUR, ["EUR => USD", "EUR => CZK", "EUR => CAD"]],
      [CZK, ["CZK => EUR", "CZK => EUR => CAD", "CZK => EUR => USD"]],
    ])(`%#: %s conversions`, (c, exp) => {
      expect([...c.conversions.values()].map((c) => c.direction)).toEqual(exp);
    });
  });
});
