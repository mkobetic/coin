import * as d3 from "d3";

// MODELS

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
    readonly decimals: number
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
    readonly value: Amount
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

export class Account {
  children: Account[] = [];
  postings: Posting[] = [];
  constructor(
    readonly name: string,
    readonly fullName: string,
    readonly commodity: Commodity,
    readonly parent: Account,
    readonly closed?: Date
  ) {
    if (parent) {
      parent.children.push(this);
    }
  }
  allChildren(): Account[] {
    return this.children.concat(
      this.children.map((c) => c.allChildren()).flat()
    );
  }
  toString(): string {
    return this.fullName;
  }
  // child name with this account's name prefix stripped
  relativeName(child: Account): string {
    return child.fullName.slice(this.fullName.length);
  }
  withAllChildPostings(from: Date, to: Date): Posting[] {
    const postings = trimToDateRange(this.postings, from, to).concat(
      this.children.map((c) => c.withAllChildPostings(from, to)).flat()
    );
    postings.sort(
      (a, b) => a.transaction.posted.getTime() - b.transaction.posted.getTime()
    );
    return postings;
  }
  withAllChildPostingGroups(
    from: Date,
    to: Date,
    groupKey: d3.TimeInterval
  ): AccountPostingGroups[] {
    let accounts = this.allChildren();
    accounts.unshift(this);
    // drop accounts with no postings
    accounts = accounts.filter((a) => a.postings.length > 0);
    return accounts.map((acc) => ({
      account: acc,
      groups: groupBy(
        trimToDateRange(acc.postings, from, to),
        groupKey,
        (p) => p.transaction.posted,
        acc.commodity
      ),
    }));
  }
  getRootAccount(): Account {
    return this.parent ? this.parent.getRootAccount() : this;
  }
}

export interface Tags {
  [key: string]: string;
}

export class Posting {
  index?: number; // used to cache index in the register view for sorting
  constructor(
    readonly transaction: Transaction,
    readonly account: Account,
    readonly quantity: Amount,
    readonly balance: Amount,
    readonly balance_asserted?: boolean,
    readonly notes?: string[],
    readonly tags?: Tags
  ) {
    transaction.postings.push(this);
    account.postings.push(this);
  }
  toString(): string {
    return (
      this.account.fullName +
      " " +
      this.quantity.toString() +
      (this.balance_asserted ? " = " + this.balance.toString() : "")
    );
  }
}

export class Transaction {
  postings: Posting[] = [];
  constructor(
    readonly posted: Date,
    readonly description: string,
    readonly notes?: string[],
    readonly code?: string
  ) {}
  toString(): string {
    return dateToString(this.posted) + " " + this.description;
  }
  // return the other posting in this transaction
  // less clear in multi-posting transactions
  // return first posting that isn't notThis and has the opposite sign
  other(notThis: Posting): Posting {
    const notThisSign = notThis.quantity.sign;
    for (const p of this.postings) {
      if (p != notThis && (p.quantity.sign != notThisSign || notThisSign == 0))
        return p;
    }
    throw new Error(`No other posting? ${notThis}`);
  }
}

// IMPORT ALL DATA

type importedCommodities = Record<
  string,
  { id: string; name: string; decimals: number }
>;
type importedPrices = {
  commodity: string;
  currency: string;
  time: string;
  value: string;
}[];

const Commodities: Record<string, Commodity> = {};
function loadCommodities() {
  const importedCommodities = JSON.parse(
    document.getElementById("importedCommodities")!.innerText
  ) as importedCommodities;
  for (const impCommodity of Object.values(importedCommodities)) {
    const commodity = new Commodity(
      impCommodity.id,
      impCommodity.name,
      impCommodity.decimals
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
      const price = new Price(commodity, new Date(imported.time), amount);
      commodity.prices.push(price);
    }
  }
}

type importedAccounts = Record<
  string,
  {
    name: string;
    fullName: string;
    commodity: string;
    parent: string;
    closed?: string;
  }
>;
type importedTransactions = {
  posted: string;
  description: string;
  postings: {
    account: string;
    balance: string;
    balance_asserted: boolean;
    quantity: string;
    notes?: string[];
    tags?: Tags;
  }[];
  notes?: string[];
  code?: string;
  tags?: Tags;
}[];

// min and max transaction date from the dataset
export let MinDate = new Date();
export let MaxDate = new Date(0);

export const Accounts: Record<string, Account> = {};
export const Roots: Account[] = [];
function loadAccounts() {
  const importedAccounts = JSON.parse(
    document.getElementById("importedAccounts")!.innerText
  ) as importedAccounts;
  for (const impAccount of Object.values(importedAccounts)) {
    if (impAccount.name == "Root") continue;
    const parent = Accounts[impAccount.parent];
    const account = new Account(
      impAccount.name,
      impAccount.fullName,
      Commodities[impAccount.commodity],
      parent,
      impAccount.closed ? new Date(impAccount.closed) : undefined
    );
    Accounts[account.fullName] = account;
    if (!parent) {
      Roots.push(account);
    }
  }

  const importedTransactions = JSON.parse(
    document.getElementById("importedTransactions")!.innerText
  ) as importedTransactions;
  for (const impTransaction of Object.values(importedTransactions)) {
    const posted = new Date(impTransaction.posted);
    if (posted < MinDate) MinDate = posted;
    if (MaxDate < posted) MaxDate = posted;
    const transaction = new Transaction(
      posted,
      impTransaction.description,
      impTransaction.notes,
      impTransaction.code
    );
    for (const impPosting of impTransaction.postings) {
      const account = Accounts[impPosting.account];
      if (!account) {
        throw new Error("Unknown account: " + impPosting.account);
      }
      const quantity = Amount.parse(impPosting.quantity);
      const balance = Amount.parse(impPosting.balance);
      const posting = new Posting(
        transaction,
        account,
        quantity,
        balance,
        impPosting.balance_asserted,
        impPosting.notes,
        impPosting.tags
      );
    }
  }
  MinDate = new Date(MinDate.getFullYear(), 0, 1);
  MaxDate = new Date(MaxDate.getFullYear(), 11, 31);
}

export function loadEverything() {
  loadCommodities();
  loadAccounts();
}

// UTILS

export function dateToString(date: Date): string {
  return date.toISOString().split("T")[0];
}

export function trimToDateRange(postings: Posting[], start: Date, end: Date) {
  const from = postings.findIndex((p) => p.transaction.posted >= start);
  if (from < 0) return [];
  const to = postings.findIndex((p) => p.transaction.posted > end);
  if (to < 0) return postings.slice(from);
  return postings.slice(from, to);
}

// single entry of a list of postings grouped by some key (week,month,...)
export type PostingGroup = {
  date: Date;
  postings: Posting[];
  sum: Amount; // sum of posting amounts
  total: Amount; // running total across an array of groups
  offset?: number; // used to cache offset value (x) in layered stack chart
  width?: number; // used to cache width value (x) in layered stack chart
};

export function groupBy(
  postings: Posting[],
  groupBy: d3.TimeInterval,
  date: (p: Posting) => Date,
  commodity: Commodity
): PostingGroup[] {
  const groups = new Map<string, Posting[]>();
  for (const p of postings) {
    const k = dateToString(groupBy(date(p)));
    const group = groups.get(k);
    group ? group.push(p) : groups.set(k, [p]);
  }
  const data: PostingGroup[] = [];
  const total = new Amount(0, commodity);
  return groupBy.range(State.StartDate, State.EndDate).map((date) => {
    let postings = groups.get(dateToString(date));
    const sum = new Amount(0, commodity);
    if (!postings) {
      postings = [];
    } else {
      postings.forEach((p) => sum.addIn(p.quantity, date));
      total.addIn(sum, date);
    }
    return { date, postings, sum, total: Amount.clone(total) };
  });
}

// Take an array of account posting groups and total them all by
// adding the rest into the first one, return the first
function addIntoFirst(groups: AccountPostingGroups[]): AccountPostingGroups {
  // DESTRUCTIVE! add everything up into the first group
  const total = groups[0];
  const rest = groups.slice(1);
  total.groups.forEach((g, i) => {
    rest.forEach((gs) => {
      const g2 = gs.groups[i];
      if (g.date.getTime() != g2.date.getTime())
        throw new Error("date mismatch totaling groups");
      g.postings.push(...g2.postings);
      g.sum.addIn(g2.sum, g.date);
      g.total.addIn(g2.total, g.date);
    });
  });
  total.account = undefined;
  return total;
}

// list of groups for an account
export type AccountPostingGroups = {
  account?: Account;
  groups: PostingGroup[];
};

export function groupWithSubAccounts(
  account: Account,
  groupKey: d3.TimeInterval,
  maxAccounts: number,
  options?: {
    negated?: boolean;
  }
) {
  const opts = { negated: false }; // default
  Object.assign(opts, options);
  // get all account group lists
  const groups = account.withAllChildPostingGroups(
    State.StartDate,
    State.EndDate,
    groupKey
  );
  // compute average for each account
  const averages = groups.map((g, i) => {
    const postings = g.groups;
    return {
      index: i,
      avg: postings[postings.length - 1].total.toNumber() / postings.length,
    };
  });
  // sort by average and pick top accounts
  averages.sort((a, b) => (opts.negated ? a.avg - b.avg : b.avg - a.avg));
  const top = averages.slice(0, maxAccounts).map((avg) => groups[avg.index]);
  // if there's more accounts than maxAccounts, total the rest into an Other list
  if (averages.length > maxAccounts) {
    // total the rest into other
    const other = addIntoFirst(
      averages.slice(maxAccounts - 1).map((avg) => groups[avg.index])
    );
    // replace last with other
    top.pop();
    top.push(other);
  }
  return top;
}

// UI

export const Aggregation = {
  None: null,
  Weekly: d3.timeWeek,
  Monthly: d3.timeMonth,
  Quarterly: d3.timeMonth.every(3),
  Yearly: d3.timeYear,
};

// UI State
export const State = {
  // All these need to be set again after loadEverything() is called
  SelectedAccount: Accounts.Assets,
  SelectedView: "Register",
  StartDate: MinDate,
  EndDate: MaxDate,
  View: {
    // Should we recurse into subaccounts
    ShowSubAccounts: false,
    ShowNotes: false, // Show notes in register view
    Aggregate: "None" as keyof typeof Aggregation,
    // How many largest subaccounts to show when aggregating.
    AggregatedSubAccountMax: 5,
  },
};
