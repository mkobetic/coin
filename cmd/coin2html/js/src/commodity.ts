import { scaleTime } from "d3-scale";
import { timeWeek } from "d3-time";
import { dateToString, last } from "./utils";

// Commodity, Amount and Price

// Conversion produces price from date.
type Conversion = {
  (d: Date): Price;
  from: Commodity;
  to: Commodity;
  direction: string;
  steps: number;
};

export function newConversion(knownPrices: Price[]): Conversion {
  // knownPrices are assumed to be sorted by time
  if (knownPrices.length == 0)
    throw new Error("cannot create conversion from empty price list");
  const from = knownPrices[0].date;
  const to = last(knownPrices)!.date;
  const dates = timeWeek.range(from, to);
  const params = {
    from: knownPrices[0].commodity,
    to: knownPrices[0].value.commodity,
    dates,
    direction: knownPrices[0].direction,
    steps: 1,
  };
  if (dates.length == 0) {
    return Object.assign((d: Date) => knownPrices[0], params);
  }
  // scale from dates to the index of the week in the date range
  const scale = scaleTime([from, to], [0, dates.length - 1]).clamp(true);
  // generate array of weekly prices
  let cpi = 0;
  const weeklyPrices = dates.map((d) => {
    while (knownPrices[cpi].date < d) cpi++;
    return knownPrices[cpi];
  });
  // conversion function, add closed over elements as properties for debugging
  const conversion = (d: Date) => weeklyPrices[Math.round(scale(d))];

  return Object.assign(conversion, {
    scale,
    weeks: weeklyPrices,
    ...params,
  });
}

export function composeConversions(
  conversion: Conversion,
  conversion2: Conversion
) {
  if (conversion.steps > 1 || conversion.to != conversion2.from)
    throw new Error(
      `cannot compose conversions ${conversion.direction} and ${conversion2.direction}`
    );
  return Object.assign(
    (d: Date) =>
      new Price(
        conversion.from,
        d,
        conversion(d).value.convertTo(conversion2(d)),
        ""
      ),
    {
      from: conversion.from,
      to: conversion2.to,
      direction: conversion.from.toString() + " => " + conversion2.direction,
      conversion,
      conversion2,
      steps: conversion.steps + conversion2.steps,
    }
  );
}

export class Commodity {
  prices: Price[] = [];
  _conversions?: Map<Commodity, Conversion>;

  constructor(
    readonly id: string,
    readonly name: string,
    readonly decimals: number,
    readonly location?: string
  ) {}

  static find(id: string): Commodity {
    const c = Commodities[id];
    if (!c) throw new Error(`unknown commodity ${id}`);
    return c;
  }
  toString(): string {
    return this.id;
  }
  // conversion functions created from prices, by price commodity
  // needs to be lazy because prices are added during loading
  get conversions() {
    if (this._conversions) return this._conversions;
    // group prices by price commodity
    const pricesByCommodity = new Map<Commodity, Price[]>();
    // make sure the prices are sorted correctly
    this.prices.sort((a, b) => a.date.getTime() - b.date.getTime());
    this.prices.forEach((p) => {
      const cps = pricesByCommodity.get(p.value.commodity);
      if (cps) cps.push(p);
      else pricesByCommodity.set(p.value.commodity, [p]);
    });
    // build a conversion function for each price commodity
    this._conversions = new Map();
    for (const [commodity, cps] of pricesByCommodity) {
      const conversion = newConversion(cps);
      this._conversions.set(commodity, conversion);
    }
    return this._conversions;
  }
  findConversion(to: Commodity): Conversion | undefined {
    function breadthFirstSearch(
      queue: [Commodity, Conversion[]][], // path of conversions in reverse order
      visited: Set<Commodity> // visited commodities
    ): Conversion[] | undefined {
      while (queue.length > 0) {
        const [commodity, path] = queue.shift()!;
        const conversion = commodity.conversions.get(to);
        // cannot use multi-step conversions in the search because that won't yield shortest path
        if (conversion && conversion.steps == 1) return [conversion, ...path];
        for (const [commodity2, conversion] of commodity.conversions) {
          if (conversion.steps > 1 || visited.has(commodity2)) continue;
          queue.push([commodity2, [conversion, ...path]]);
          visited.add(commodity2);
        }
      }
      return undefined;
    }
    const path = breadthFirstSearch([[this, []]], new Set([this]));
    if (!path) return undefined;
    if (path.length == 1) return path[0];
    return path.slice(1).reduce((previous, conversion) => {
      const composed = composeConversions(conversion, previous);
      composed.from._conversions!.set(composed.to, composed);
      return composed;
    }, path[0]);
  }

  // convert amount to this commodity using price on given date
  convert(amount: Amount, date: Date): Amount {
    if (amount.commodity == this) return amount;
    if (amount.isZero) return new Amount(0, this);
    const conversion = amount.commodity.findConversion(this);
    if (!conversion)
      throw new Error(
        `Cannot convert ${amount.toString()} to ${this.toString()}`
      );
    const price = conversion(date);
    return amount.convertTo(price);
  }
}

export const Commodities: Record<string, Commodity> = {};
// template parser, e.g. commodity`CAD`
export function commodity(strings: TemplateStringsArray): Commodity {
  if (strings.length != 1)
    throw new Error(`invalid commodity template ${strings}`);
  return Commodity.find(strings[0]);
}

export class Amount {
  constructor(private value: number, readonly commodity: Commodity) {}
  static clone(amount: Amount) {
    return new Amount(amount.value, amount.commodity);
  }
  static parse(input: string): Amount {
    const parts = input.split(" ");
    if (parts.length != 2) {
      throw new Error("Invalid amount: " + input);
    }
    const commodity = Commodity.find(parts[1]);

    // drop the decimal point, make sure value aligns with commodity.decimals.
    let [int, dec] = parts[0].split(".");
    if (commodity.decimals > 0) {
      if (!dec) dec = "";
      dec =
        dec.length > commodity.decimals
          ? dec.slice(0, commodity.decimals)
          : dec + "0".repeat(commodity.decimals - dec.length);
      // drop thousands separators if present
      int = int.replace(",", "") + dec;
    }
    const value = parseInt(int);
    return new Amount(value, commodity);
  }
  toString(thousandsSeparator = true): string {
    let str = Math.abs(this.value).toString();
    if (this.commodity.decimals > 0) {
      if (str.length < this.commodity.decimals) {
        str = "0".repeat(this.commodity.decimals - str.length + 1) + str;
      }
      const intPart = str.slice(0, -this.commodity.decimals);
      str =
        (thousandsSeparator ? triplets(intPart).join(",") : intPart) +
        "." +
        str.slice(-this.commodity.decimals);
      if (str[0] == ".") {
        str = "0" + str;
      }
    }
    return (this.value < 0 ? "-" : "") + str + " " + this.commodity.id;
  }
  toNumber() {
    return this.value / 10 ** this.commodity.decimals;
  }
  addIn(amount: Amount, date: Date): Amount {
    if (amount.commodity == this.commodity) {
      this.value += amount.value;
      return this;
    }
    return this.addIn(this.commodity.convert(amount, date), date);
  }
  convertTo(price: Price): Amount {
    // the product decimals is a sum of this and price decimals, so divide by this decimals
    const float =
      (this.value * price.value.value) / 10 ** this.commodity.decimals;
    // accounting rounding should round 0.5 up
    return new Amount(Math.round(float), price.value.commodity);
  }
  cmp(amount: Amount, absolute = false) {
    if (this.commodity != amount.commodity) {
      throw new Error("comparing different commodities");
    }
    return absolute
      ? Math.abs(this.value) - Math.abs(amount.value)
      : this.value - amount.value;
  }
  reciprocal(decimals: number): number {
    const reciprocal = 10 ** this.commodity.decimals / this.value;
    return Math.round(reciprocal * 10 ** decimals);
  }
  get sign() {
    return Math.sign(this.value);
  }
  get isZero() {
    return this.value == 0;
  }
}

// template parser, e.g. amount`10.00 CAD`
export function amount(strings: TemplateStringsArray): Amount {
  if (strings.length != 1)
    throw new Error(`invalid amount template ${strings}`);
  return Amount.parse(strings[0]);
}

export class Price {
  constructor(
    readonly commodity: Commodity,
    readonly date: Date,
    readonly value: Amount,
    readonly location?: string
  ) {}
  static parse(input: string): Price {
    const parts = input.split(":");
    if (parts.length != 2) {
      throw new Error("Invalid price: " + input);
    }
    const commodity = Commodity.find(parts[0].trim());
    const [value, dateString] = parts[1].split("@");
    const date = dateString ? new Date(dateString.trim()) : new Date();
    return new Price(commodity, date, Amount.parse(value.trim()), "unknown");
  }

  toString(): string {
    return (
      this.commodity.toString() +
      ": " +
      this.value.toString() +
      " @ " +
      dateToString(this.date)
    );
  }
  reverse(): Price {
    return new Price(
      this.value.commodity,
      this.date,
      new Amount(
        this.value.reciprocal(this.commodity.decimals),
        this.commodity
      ),
      this.location
    );
  }
  get direction() {
    return this.commodity.toString() + " => " + this.value.commodity.toString();
  }
}

// template parser, e.g. price`USD: 1.33 CAD`
export function price(strings: TemplateStringsArray): Price {
  if (strings.length != 1)
    throw new Error(`invalid amount template ${strings}`);
  return Price.parse(strings[0]);
}

type importedCommodities = Record<
  string,
  { id: string; name: string; decimals: number; location: string }
>;
type importedPrices = {
  commodity: string;
  time: string;
  value: string;
  location: string;
}[];

export function loadCommodities(source: string) {
  const importedCommodities = JSON.parse(source) as importedCommodities;
  for (const impCommodity of Object.values(importedCommodities)) {
    const commodity = new Commodity(
      impCommodity.id,
      impCommodity.name,
      impCommodity.decimals,
      impCommodity.location
    );
    Commodities[commodity.id] = commodity;
  }
}

export function loadPrices(source: string) {
  const importedPrices = JSON.parse(source) as importedPrices;
  if (importedPrices) {
    for (const imported of importedPrices) {
      const commodity = Commodity.find(imported.commodity);
      const amount = Amount.parse(imported.value);
      if (amount.toString(false) != imported.value) {
        throw new Error(
          `Parsed amount "${amount}" doesn't match imported "${imported.value}"`
        );
      }
      const price = new Price(
        commodity,
        new Date(imported.time),
        amount,
        imported.location
      );
      commodity.prices.push(price);
      amount.commodity.prices.push(price.reverse());
    }
  }
}

function triplets(s: string): string[] {
  const triplets = [];
  for (let end = s.length; end > 0; end = end - 3) {
    let start = end - 3;
    if (start < 0) start = 0;
    triplets.unshift(s.slice(start, end));
  }
  return triplets;
}
