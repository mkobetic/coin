import * as d3 from "d3";
import { dateToString } from "./utils";

// Commodity, Amount and Price

// Conversion produces price from date.
type Conversion = (d: Date) => Amount;

function newConversion(prices: Price[]): Conversion {
  if (prices.length == 0)
    throw new Error("cannot create conversion from empty price list");
  const from = prices[0].date;
  const to = prices[prices.length - 1].date;
  const dates = d3.timeWeek.range(from, to);
  if (dates.length == 0) return (d: Date) => prices[0].value;
  // scale from dates to the number of weeks/price points
  const scale = d3.scaleTime([from, to], [0, dates.length - 1]).clamp(true);
  // generate array of prices per week
  let cpi = 0;
  const weeks = dates.map((d) => {
    while (prices[cpi].date < d) cpi++;
    return prices[cpi].value;
  });
  // conversion function, add closed over elements as properties for debugging
  const conversion = (d: Date) => weeks[Math.round(scale(d))];
  conversion.scale = scale;
  conversion.weeks = weeks;
  conversion.dates = dates;
  return conversion;
}

export class Commodity {
  prices: Price[] = [];
  _conversions?: Map<Commodity, Conversion>;
  constructor(
    readonly id: string,
    readonly name: string,
    readonly decimals: number,
    readonly location: string
  ) {}
  toString(): string {
    return this.id;
  }
  // conversion functions created from prices, by price commodity
  // needs to be lazy because prices are added during loading
  get conversions() {
    if (this._conversions) return this._conversions;
    // group prices by price commodity
    const prices = new Map<Commodity, Price[]>();
    this.prices.forEach((p) => {
      const cps = prices.get(p.value.commodity);
      if (cps) cps.push(p);
      else prices.set(p.value.commodity, [p]);
    });
    // build a conversion function for each price commodity
    this._conversions = new Map();
    for (const [commodity, cps] of prices) {
      const conversion = newConversion(cps);
      this._conversions.set(commodity, conversion);
    }
    return this._conversions;
  }
  // convert amount to this commodity using price on given date
  convert(amount: Amount, date: Date): Amount {
    if (amount.isZero) return new Amount(0, this);
    if (amount.commodity == this) return amount;
    const conversion = amount.commodity.conversions.get(this);
    if (!conversion)
      throw new Error(
        `Cannot convert ${amount.toString()} to ${this.toString()}`
      );
    const price = conversion(date);
    return amount.convertTo(price);
  }
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
    const commodity = Commodities[parts[1]];
    if (!commodity) {
      throw new Error("Unknown commodity: " + parts[1]);
    }
    // drop the decimal point, commodity.decimals should indicate where it is.
    const value = Number(parts[0].replace(".", ""));
    return new Amount(value, commodity);
  }
  toString(): string {
    let str = this.value.toString();
    if (this.commodity.decimals > 0) {
      if (str.length < this.commodity.decimals) {
        str = "0".repeat(this.commodity.decimals - str.length + 1) + str;
      }
      str =
        str.slice(0, -this.commodity.decimals) +
        "." +
        str.slice(-this.commodity.decimals);
      if (str[0] == ".") {
        str = "0" + str;
      }
    }
    return str + " " + this.commodity.id;
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
  convertTo(price: Amount): Amount {
    // the product decimals is a sum of this and price decimals, so divide by this decimals
    const float = (this.value * price.value) / 10 ** this.commodity.decimals;
    // accounting rounding should round 0.5 up
    return new Amount(Math.round(float), price.commodity);
  }
  cmp(amount: Amount) {
    const decimalDiff = this.commodity.decimals - amount.commodity.decimals;
    return decimalDiff < 0
      ? this.value * 10 ** -decimalDiff - amount.value
      : this.value - amount.value * 10 ** decimalDiff;
  }
  get sign() {
    return Math.sign(this.value);
  }
  get isZero() {
    return this.value == 0;
  }
}

export class Price {
  constructor(
    readonly commodity: Commodity,
    readonly date: Date,
    readonly value: Amount,
    readonly location: string
  ) {
    commodity.prices.push(this);
  }
  toString(): string {
    return (
      this.commodity.toString() +
      ": " +
      this.value.toString() +
      "@" +
      dateToString(this.date)
    );
  }
}

type importedCommodities = Record<
  string,
  { id: string; name: string; decimals: number; location: string }
>;
type importedPrices = {
  commodity: string;
  currency: string;
  time: string;
  value: string;
  location: string;
}[];

export const Commodities: Record<string, Commodity> = {};
export function loadCommodities() {
  const importedCommodities = JSON.parse(
    document.getElementById("importedCommodities")!.innerText
  ) as importedCommodities;
  for (const impCommodity of Object.values(importedCommodities)) {
    const commodity = new Commodity(
      impCommodity.id,
      impCommodity.name,
      impCommodity.decimals,
      impCommodity.location
    );
    Commodities[commodity.id] = commodity;
  }

  const importedPrices = JSON.parse(
    document.getElementById("importedPrices")!.innerText
  ) as importedPrices;
  if (importedPrices) {
    for (const imported of importedPrices) {
      const commodity = Commodities[imported.commodity];
      if (!commodity) {
        throw new Error("Unknown commodity: " + imported.commodity);
      }
      const amount = Amount.parse(imported.value);
      if (amount.toString() != imported.value) {
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
    }
  }
}
