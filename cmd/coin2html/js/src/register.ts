import {
  Aggregation,
  State,
  addAggregateInput,
  addIncludeNotesInput,
  addIncludeSubAccountsInput,
  addSubAccountMaxInput,
  emptyElement,
  MainView,
  addShowLocationInput,
  addAggregationStyleInput,
  AggregationStyle,
} from "./views";
import { Account, Posting } from "./account";
import {
  dateToString,
  groupBy,
  groupWithSubAccounts,
  last,
  trimToDateRange,
} from "./utils";
import { Amount } from "./commodity";
import { select } from "d3-selection";

function addTableWithHeader(containerSelector: string, labels: string[]) {
  const table = select(containerSelector)
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

export function viewRegister(options?: {
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
  if (State.View.Aggregate == "None") {
    addIncludeNotesInput(containerSelector);
    addShowLocationInput(containerSelector);
  } else {
    addAggregationStyleInput(containerSelector);
    if (State.View.ShowSubAccounts) addSubAccountMaxInput(containerSelector);
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
        [
          State.View.AggregationStyle == AggregationStyle.Balances
            ? g.balance
            : g.sum,
          "amount",
        ],
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
    const balance = new Amount(0, account.commodity);
    const sum = new Amount(0, account.commodity);
    const postings: Posting[] = [];
    const row = groups.map((gs) => {
      const g = gs.groups[i];
      if (g.date.getTime() != date.getTime())
        throw new Error("date mismatch transposing groups");
      postings.push(...g.postings);
      sum.addIn(g.sum, g.date);
      balance.addIn(g.balance, g.date);
      return g;
    });
    total.addIn(sum, date);
    row.push({
      date: date,
      postings,
      sum,
      total: Amount.clone(total),
      balance,
    });
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
      const total = last(row)!;
      const columns = row.map((g) => [
        State.View.AggregationStyle == AggregationStyle.Flows
          ? g.sum
          : g.balance,
        "amount",
      ]);
      // prepend date
      columns.unshift([dateToString(row[0].date), "date"]);
      // append total correctly
      if (options.aggregatedTotal)
        columns.push([
          State.View.AggregationStyle == AggregationStyle.Flows
            ? total.total
            : total.balance,
          "amount",
        ]);
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
  const labels = [
    "Date",
    "Description",
    "Account",
    "Amount",
    "Balance",
    "Cum.Total",
  ];
  if (State.View.ShowLocation) labels.push("Location");
  const table = addTableWithHeader(containerSelector, labels);
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
      const values = [
        [dateToString(p.transaction.posted), "date"],
        [p.transaction.description, "text"],
        [p.transaction.other(p).account, "account"],
        [p.quantity, "amount"],
        [p.balance, "amount"],
        [Amount.clone(total), "amount"],
      ];
      if (State.View.ShowLocation)
        values.push([p.transaction.location, "text"]);
      return values;
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
  const labels = [
    "Date",
    "Description",
    "SubAccount",
    "Account",
    "Amount",
    "Cum.Total",
  ];
  if (State.View.ShowLocation) labels.push("Location");
  const table = addTableWithHeader(containerSelector, labels);
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
      const values = [
        [dateToString(p.transaction.posted), "date"],
        [p.transaction.description, "text"],
        [account.relativeName(p.account), "account"],
        [p.transaction.other(p).account, "account"],
        [p.quantity, "amount"],
        [Amount.clone(total), "amount"],
      ];
      if (State.View.ShowLocation)
        values.push([p.transaction.location, "text"]);
      return values;
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
