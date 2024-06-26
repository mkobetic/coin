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

class Commodity {
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

class Amount {
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

class Price {
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

class Account {
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

interface Tags {
  [key: string]: string;
}

class Posting {
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

class Transaction {
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
let MinDate = new Date();
let MaxDate = new Date(0);

const Accounts: Record<string, Account> = {};
const Roots: Account[] = [];
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

function loadEverything() {
  loadCommodities();
  loadAccounts();
}

// Need to load before initializing the UI state below.
loadEverything();

// UTILS

function dateToString(date: Date): string {
  return date.toISOString().split("T")[0];
}

function trimToDateRange(postings: Posting[], start: Date, end: Date) {
  const from = postings.findIndex((p) => p.transaction.posted >= start);
  if (from < 0) return [];
  const to = postings.findIndex((p) => p.transaction.posted > end);
  if (to < 0) return postings.slice(from);
  return postings.slice(from, to);
}

// single entry of a list of postings grouped by some key (week,month,...)
type PostingGroup = {
  date: Date;
  postings: Posting[];
  sum: Amount; // sum of posting amounts
  total: Amount; // running total across an array of groups
  offset?: number; // used to cache offset value (x) in layered stack chart
  width?: number; // used to cache width value (x) in layered stack chart
};

function groupBy(
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
type AccountPostingGroups = { account?: Account; groups: PostingGroup[] };

function groupWithSubAccounts(
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

const Aggregation = {
  None: null,
  Weekly: d3.timeWeek,
  Monthly: d3.timeMonth,
  Quarterly: d3.timeMonth.every(3),
  Yearly: d3.timeYear,
};

// View types by account category.
// All types have Register.
const Views = {
  Assets: {
    Register: viewRegister,
    Chart: viewChart,
  },
  Liabilities: {
    Register: () => viewRegister({ negated: true }),
    Chart: () => viewChart({ negated: true }),
  },
  Income: {
    Register: () =>
      viewRegister({
        negated: true,
        aggregatedTotal: true,
      }),
    Chart: () => viewChart({ negated: true }),
  },
  Expenses: {
    Register: () =>
      viewRegister({
        aggregatedTotal: true,
      }),
    Chart: viewChart,
  },
  Equity: {
    Register: viewRegister,
  },
  Unbalanced: {
    Register: viewRegister,
  },
};

// UI State
let State = {
  SelectedAccount: Accounts.Assets,
  SelectedView: Object.keys(Views.Assets)[0],
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

// VIEWS

function addIncludeSubAccountsInput(containerSelector: string) {
  const container = d3.select(containerSelector);
  container
    .append("label")
    .property("for", "includeSubAccounts")
    .text("SubAccounts");
  container
    .append("input")
    .on("change", (e, d) => {
      const input = e.currentTarget as HTMLInputElement;
      State.View.ShowSubAccounts = input.checked;
      updateView();
    })
    .attr("id", "includeSubAccounts")
    .attr("type", "checkbox")
    .property("checked", State.View.ShowSubAccounts);
}

function addIncludeNotesInput(containerSelector: string) {
  const container = d3.select(containerSelector);
  container.append("label").property("for", "includeNotes").text("Show Notes");
  container
    .append("input")
    .on("change", (e, d) => {
      const input = e.currentTarget as HTMLInputElement;
      State.View.ShowNotes = input.checked;
      updateView();
    })
    .attr("id", "includeNotes")
    .attr("type", "checkbox")
    .property("checked", State.View.ShowNotes);
}

function addSubAccountMaxInput(containerSelector: string) {
  const container = d3.select(containerSelector);
  container
    .append("label")
    .property("for", "subAccountMax")
    .text("SubAccount Max");
  container
    .append("input")
    .on("change", (e, d) => {
      const input = e.currentTarget as HTMLInputElement;
      State.View.AggregatedSubAccountMax = parseInt(input.value);
      updateView();
    })
    .attr("id", "subAccountMax")
    .attr("type", "number")
    .property("value", State.View.AggregatedSubAccountMax);
}

function addAggregateInput(
  containerSelector: string,
  options?: {
    includeNone?: boolean;
  }
) {
  const opts = { includeNone: true }; // defaults
  Object.assign(opts, options);
  const container = d3.select(containerSelector);
  container.append("label").property("for", "aggregate").text("Aggregate");
  const aggregate = container.append("select").attr("id", "aggregate");
  aggregate.on("change", (e, d) => {
    const select = e.currentTarget as HTMLSelectElement;
    const selected = select.options[select.selectedIndex].value;
    State.View.Aggregate = selected as keyof typeof Aggregation;
    updateView();
  });
  let data = Object.keys(Aggregation).filter(
    (k) => opts.includeNone || k != "None"
  );
  if (!opts.includeNone && State.View.Aggregate == "None") {
    State.View.Aggregate = data[0] as keyof typeof Aggregation;
    console.log("Aggregate = ", State.View.Aggregate);
  }
  aggregate
    .selectAll("option")
    .data(data)
    .join("option")
    .property("selected", (v) => v == State.View.Aggregate)
    .property("value", (v) => v)
    .text((v) => v);
}

// REGISTER

function addTableWithHeader(containerSelector: string, labels: string[]) {
  const table = d3
    .select(containerSelector)
    .append("table")
    .attr("id", "register");
  table
    .append("thead")
    .append("tr")
    .selectAll("th")
    .data(labels)
    .join("th")
    .text((d) => d);
  return table;
}

function viewRegister(options?: {
  negated?: boolean; // is this negatively denominated account (e.g. Income/Liability)
  aggregatedTotal?: boolean; // include cumulative total in aggregated register
}) {
  const containerSelector = MainView;
  const account = State.SelectedAccount;
  const opts = { negated: false, aggregatedTotal: false };
  Object.assign(opts, options);
  // clear out the container
  emptyElement(containerSelector);
  addIncludeSubAccountsInput(containerSelector);
  addAggregateInput(containerSelector);
  if (State.View.ShowSubAccounts && State.View.Aggregate != "None")
    addSubAccountMaxInput(containerSelector);
  if (State.View.Aggregate == "None") {
    addIncludeNotesInput(containerSelector);
  }
  const groupKey = Aggregation[State.View.Aggregate];
  if (groupKey) {
    if (State.View.ShowSubAccounts)
      viewRegisterAggregatedWithSubAccounts(
        containerSelector,
        groupKey,
        account,
        opts
      );
    else viewRegisterAggregated(containerSelector, groupKey, account, opts);
  } else {
    if (State.View.ShowSubAccounts)
      viewRegisterFullWithSubAccounts(containerSelector, account, opts);
    else viewRegisterFull(containerSelector, account, opts);
  }
}

function viewRegisterAggregated(
  containerSelector: string,
  groupKey: d3.TimeInterval,
  account: Account,
  options: {
    negated: boolean;
    aggregatedTotal: boolean;
  }
) {
  const labels = ["Date", "Amount"];
  if (options.aggregatedTotal) labels.push("Cum.Total");
  const table = addTableWithHeader(containerSelector, labels);
  const data = groupBy(
    account.postings,
    groupKey,
    (p) => p.transaction.posted,
    account.commodity
  );
  table
    .append("tbody")
    .selectAll("tr")
    .data(data)
    .join("tr")
    .classed("even", (_, i) => i % 2 == 0)
    .selectAll("td")
    .data((g) => {
      const row = [
        [dateToString(g.date), "date"],
        [g.sum, "amount"],
      ];
      if (options.aggregatedTotal) row.push([g.total, "amount"]);
      return row;
    })
    .join("td")
    .classed("amount", ([v, c]) => c == "amount")
    .text(([v, c]) => v.toString());
}

function viewRegisterAggregatedWithSubAccounts(
  containerSelector: string,
  groupKey: d3.TimeInterval,
  account: Account,
  options: {
    negated: boolean;
    aggregatedTotal: boolean;
  }
) {
  const dates = groupKey.range(State.StartDate, State.EndDate);
  const groups = groupWithSubAccounts(
    account,
    groupKey,
    State.View.AggregatedSubAccountMax,
    options
  );
  // transpose the groups into row data
  const total = new Amount(0, account.commodity);
  const data = dates.map((date, i) => {
    const sum = new Amount(0, account.commodity);
    const postings: Posting[] = [];
    const row = groups.map((gs) => {
      const g = gs.groups[i];
      if (g.date.getTime() != date.getTime())
        throw new Error("date mismatch transposing groups");
      postings.push(...g.postings);
      sum.addIn(g.sum, g.date);
      return g;
    });
    total.addIn(sum, date);
    row.push({ date: date, postings, sum, total: Amount.clone(total) });
    return row;
  });
  const labels = [
    "Date",
    ...groups.map((g) =>
      g.account ? account.relativeName(g.account) : "Other"
    ),
    "Total",
  ];
  if (options.aggregatedTotal) labels.push("Cum.Total");
  const table = addTableWithHeader(containerSelector, labels);
  table
    .append("tbody")
    .selectAll("tr")
    .data(data)
    .join("tr")
    .classed("even", (_, i) => i % 2 == 0)
    .selectAll("td")
    .data((row) => {
      const total = row[row.length - 1];
      const columns = row.map((g) => [g.sum, "amount"]);
      // prepend date
      columns.unshift([dateToString(row[0].date), "date"]);
      // append total correctly
      if (options.aggregatedTotal) columns.push([total.total, "amount"]);
      return columns;
    })
    .join("td")
    .classed("amount", ([v, c]) => c == "amount")
    .text(([v, c]) => v.toString());
}

function viewRegisterFull(
  containerSelector: string,
  account: Account,
  options: {
    negated: boolean;
  }
) {
  const table = addTableWithHeader(containerSelector, [
    "Date",
    "Description",
    "Account",
    "Amount",
    "Balance",
    "Cum.Total",
  ]);
  const total = new Amount(0, account.commodity);
  const data = trimToDateRange(
    account.postings,
    State.StartDate,
    State.EndDate
  );
  const rows = table.append("tbody").selectAll("tr").data(data).enter();
  rows
    .append("tr")
    .classed("even", (_, i) => i % 2 == 0)
    .selectAll("td")
    .data((p, i) => {
      p.index = i;
      total.addIn(p.quantity, p.transaction.posted);
      return [
        [dateToString(p.transaction.posted), "date"],
        [p.transaction.description, "text"],
        [p.transaction.other(p).account, "account"],
        [p.quantity, "amount"],
        [p.balance, "amount"],
        [Amount.clone(total), "amount"],
      ];
    })
    .join("td")
    .classed("amount", ([v, c]) => c == "amount")
    .attr("rowspan", (_, i) => (i == 0 && State.View.ShowNotes ? 2 : null))
    .text(([v, c]) => v.toString());
  if (State.View.ShowNotes) {
    rows
      .append("tr")
      .classed("even", (_, i) => i % 2 == 0)
      .selectAll("td")
      .data((p, i) => [p.transaction.notes])
      .join("td")
      .attr("colspan", 5)
      .text((notes) => (notes ? notes.join("; ") : ""));
    // need to resort the rows so that the note rows are next to the data rows
    // the index is set on the Postings with the data rows above
    table
      .select("tbody")
      .selectAll("tr")
      .sort((a: any, b: any) => a.index - b.index);
  }
}

function viewRegisterFullWithSubAccounts(
  containerSelector: string,
  account: Account,
  options: {
    negated: boolean;
  }
) {
  const table = addTableWithHeader(containerSelector, [
    "Date",
    "Description",
    "SubAccount",
    "Account",
    "Amount",
    "Cum.Total",
  ]);
  const total = new Amount(0, account.commodity);
  const data = account.withAllChildPostings(State.StartDate, State.EndDate);
  const rows = table.append("tbody").selectAll("tr").data(data).enter();
  rows
    .append("tr")
    .classed("even", (_, i) => i % 2 == 0)
    .selectAll("td")
    .data((p, i) => {
      p.index = i;
      total.addIn(p.quantity, p.transaction.posted);
      return [
        [dateToString(p.transaction.posted), "date"],
        [p.transaction.description, "text"],
        [account.relativeName(p.account), "account"],
        [p.transaction.other(p).account, "account"],
        [p.quantity, "amount"],
        [Amount.clone(total), "amount"],
      ];
    })
    .join("td")
    .classed("amount", ([v, c]) => c == "amount")
    .attr("rowspan", (_, i) => (i == 0 && State.View.ShowNotes ? 2 : null))
    .text(([v, c]) => v.toString());
  if (State.View.ShowNotes) {
    rows
      .append("tr")
      .classed("even", (_, i) => i % 2 == 0)
      .selectAll("td")
      .data((p, i) => [p.transaction.notes])
      .join("td")
      .attr("colspan", 5)
      .text((notes) => (notes ? notes.join("; ") : ""));
    // need to resort the rows so that the note rows are next to the data rows
    // the index is set on the Postings with the data rows above
    table
      .select("tbody")
      .selectAll("tr")
      .sort((a: any, b: any) => a.index - b.index);
  }
}

// CHART

function viewChart(options?: {
  negated?: boolean; // is this negatively denominated account (e.g. Income/Liability)
}) {
  const containerSelector = MainView;
  const account = State.SelectedAccount.getRootAccount();
  const opts = { negated: false }; // defaults
  Object.assign(opts, options);
  // clear out the container
  emptyElement(containerSelector);
  addAggregateInput(containerSelector, {
    includeNone: false,
  });
  addSubAccountMaxInput(containerSelector);

  const groupKey = Aggregation[State.View.Aggregate] as d3.TimeInterval;
  const dates = groupKey.range(State.StartDate, State.EndDate);
  const maxAccounts = State.View.AggregatedSubAccountMax;
  const accountGroups = groupWithSubAccounts(account, groupKey, maxAccounts, {
    negated: opts.negated,
  });
  const labelFromAccount = (a: Account | undefined) =>
    a ? account.relativeName(a) : "Other";
  const labels = accountGroups.map((gs) => labelFromAccount(gs.account));
  // compute offsets for each group left to right
  // and max width for the x domain
  let max = 0;
  dates.forEach((_, i) => {
    let offset = 0;
    accountGroups.forEach((gs) => {
      const group = gs.groups[i];
      group.offset = offset;
      let sum = Math.trunc(group.sum.toNumber());
      if (opts.negated) sum = -sum;
      group.width = sum < 0 ? 0 : sum;
      offset += group.width;
    });
    max = max < offset ? offset : max;
  });

  const rowHeight = 15,
    margin = { top: 3 * rowHeight, right: 50, bottom: 50, left: 100 },
    height = dates.length * rowHeight + margin.top + margin.bottom,
    textOffset = (rowHeight * 3) / 4;

  const svg = d3
    .select(containerSelector)
    .append("svg")
    .attr("id", "chart")
    .attr("width", "100%")
    .attr("height", height + margin.top + margin.bottom);

  let width =
    Math.max(Math.floor(svg.property("width")?.baseVal.value), 800) -
    margin.left -
    margin.right;

  var chart = svg
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  var x = d3.scaleLinear([0, max], [0, width]).nice();
  var y = d3.scaleTime([State.StartDate, State.EndDate], [0, height]);
  var z = d3.scaleOrdinal([0, maxAccounts], d3.schemeCategory10);
  var xAxis = d3.axisTop(x);
  var yAxis = d3.axisLeft(y).ticks(groupKey, "%Y/%m/%d");

  // bar layers
  var layer = chart
    .selectAll(".layer")
    .data(accountGroups)
    .enter()
    .append("g")
    .attr("class", "layer")
    .style("fill", (d, i) => z(i));

  // bars
  layer
    .selectAll("rect")
    .data((d) => d.groups)
    .enter()
    .append("rect")
    .attr("y", (d) => y(d.date))
    .attr("x", (d) => x(d.offset ?? 0))
    .attr("width", (d) => x(d.width ?? 0))
    .attr("height", rowHeight - 1)
    .on("click", (e, d) => console.log(e, d));

  // bar text
  layer
    .selectAll("text")
    .data((d) => d.groups)
    .enter()
    .append("text")
    .text((d) => {
      const v = d.width ?? 0;
      const w = (Math.log10(v) + 1) * 8;
      return v > 0 && x(v) > w ? v : "";
    })
    .attr("x", (d) => x((d.offset ?? 0) + (d.width ?? 0)) - 4)
    .attr("y", (d) => y(d.date) + textOffset);

  // axis
  chart.append("g").attr("class", "axis axis--x").call(xAxis);
  chart.append("g").attr("class", "axis axis--y").call(yAxis);

  var legend = svg
    .selectAll(".legend")
    .data(labels)
    .enter()
    .append("g")
    .attr("class", "legend")
    .attr("transform", "translate(" + margin.left + ",0)");

  var w = x.range()[1] / labels.length;

  legend
    .append("rect")
    .attr("x", (d, i) => w * i)
    .attr("y", 0)
    .attr("width", w)
    .attr("height", rowHeight - 1)
    .style("fill", (d, i) => z(i));

  legend
    .append("text")
    .text((d) => d)
    .attr("x", (d, i) => w * i + 10)
    .attr("y", textOffset);
}

// UI Node Selectors

const RootAccountSelect = "#sidebar select#root";
const AccountList = "#sidebar ul#accounts";

const ViewSelect = "#main #controls select#view";
const StartDateInput = "#main #controls input#start";
const EndDateInput = "#main #controls input#end";
const AccountOutput = "#main output#account";
const MainView = "#main section#view";

function emptyElement(selector: string) {
  (d3.select(selector).node() as Element).replaceChildren();
}

// UI Events

function updateView() {
  const account = State.SelectedAccount.getRootAccount();
  const selectedViews = Views[account.name as keyof typeof Views];
  const view = selectedViews[State.SelectedView as keyof typeof selectedViews];
  view();
}

function updateAccount() {
  const account = State.SelectedAccount;
  d3.select(AccountOutput).text(account.fullName);
  updateView();
}

function addViewSelect() {
  emptyElement(ViewSelect);
  const account = State.SelectedAccount.getRootAccount();
  const selectedViews = Object.keys(Views[account.name as keyof typeof Views]);
  if (!selectedViews.includes(State.SelectedView))
    State.SelectedView = selectedViews[0];
  d3.select(ViewSelect)
    .on("change", (e) => {
      const select = e.currentTarget as HTMLSelectElement;
      State.SelectedView = select.options[select.selectedIndex].value;
      updateView();
    })
    .selectAll("option")
    .data(selectedViews)
    .join("option")
    .property("selected", (l) => l == State.SelectedView)
    .text((l) => l);
}

type liWithAccount = HTMLLIElement & { __data__: Account };
function addAccountList() {
  const account = State.SelectedAccount;
  d3.select(AccountList)
    .selectAll("li")
    .data(account.allChildren())
    .join("li")
    .text((d) => d.fullName)
    .on("click", (e: Event) => {
      State.SelectedAccount = (e.currentTarget as liWithAccount).__data__;
      updateAccount();
    });
}

function updateAccounts() {
  addViewSelect();
  addAccountList();
  updateAccount();
}

function initializeUI() {
  const minDate = dateToString(new Date(MinDate.getFullYear(), 1, 1));
  const maxDate = dateToString(new Date(MaxDate.getFullYear() + 1, 1, 1));
  d3.select(EndDateInput)
    .property("valueAsDate", State.EndDate)
    .property("min", minDate)
    .property("max", maxDate)
    .on("change", (e) => {
      const input = e.currentTarget as HTMLInputElement;
      State.EndDate = new Date(input.value);
      updateView();
    });
  d3.select(StartDateInput)
    .property("valueAsDate", State.StartDate)
    .property("min", minDate)
    .property("max", maxDate)
    .on("change", (e) => {
      const input = e.currentTarget as HTMLInputElement;
      State.StartDate = new Date(input.value);
      updateView();
    });
  type optionWithAccount = HTMLOptionElement & { __data__: Account };
  d3.select(RootAccountSelect)
    .on("change", (e: Event) => {
      const select = e.currentTarget as HTMLSelectElement;
      const account = (
        select.options[select.selectedIndex] as optionWithAccount
      ).__data__;
      State.SelectedAccount = account;
      updateAccounts();
    })
    .selectAll("option")
    .data(Roots)
    .join("option")
    .property("selected", (d) => d == State.SelectedAccount)
    .text((d) => d.fullName);

  // trigger account selection
  updateAccounts();
}

initializeUI();
