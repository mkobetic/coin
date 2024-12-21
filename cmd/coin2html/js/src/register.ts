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
  showDetails,
} from "./views";
import { Account, Posting } from "./account";
import {
  balanceOrSum,
  dateToString,
  groupBy,
  groupByWithSubAccounts,
  last,
  PostingGroup,
  shortenAccountName,
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
      const row: [PostingGroup, (g: PostingGroup) => string, string][] = [
        [g, (g) => dateToString(g.date), "date"],
        [g, (g) => balanceOrSum(g).toString(), "amount"],
      ];
      if (options.aggregatedTotal)
        row.push([g, (g) => g.total.toString(), "amount"]);
      return row;
    })
    .join("td")
    .classed("amount", ([g, v, c]) => c == "amount")
    .text(([g, v, c]) => v(g))
    .on("click", (e, [g, v, c]) => showDetails(g));
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
  const groups = groupByWithSubAccounts(
    account,
    groupKey,
    State.View.AggregatedSubAccountMax,
    options
  );
  // convert the vertical groups into horizontal row data
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
  const maxLabelLength = Math.round(150 / State.View.AggregatedSubAccountMax);
  const labels = [
    "Date",
    ...groups.map((g) =>
      g.account
        ? shortenAccountName(account.relativeName(g.account), maxLabelLength)
        : "Other"
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
      const columns: [PostingGroup, (g: PostingGroup) => string, string][] =
        row.map((g) => [g, (g) => balanceOrSum(g).toString(), "amount"]);
      // prepend date
      columns.unshift([row[0], (g) => dateToString(g.date), "date"]);
      // append total correctly
      if (options.aggregatedTotal)
        columns.push([total, (g) => balanceOrSum(g).toString(), "amount"]);
      return columns;
    })
    .join("td")
    .classed("amount", ([g, v, c]) => c == "amount")
    .text(([g, v, c]) => v(g))
    .on("click", (e, [g, v, c]) => showDetails(g));
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
  const data = account.withAllChildPostings(State.StartDate, State.EndDate);
  renderPostingsWithSubAccounts(account, data, containerSelector, {
    showLocation: State.View.ShowLocation,
    showNotes: State.View.ShowNotes,
  });
}

export function renderPostingsWithSubAccounts(
  account: Account,
  data: Posting[],
  containerSelector: string,
  optionOverrides?: {
    showLocation?: boolean;
    showNotes?: boolean;
  }
) {
  const options = {
    showLocation: false,
    showNotes: false,
  };
  Object.assign(options, optionOverrides);
  const labels = [
    "Date",
    "Description",
    "SubAccount",
    "Account",
    "Amount",
    "Cum.Total",
  ];
  if (options.showLocation) labels.push("Location");
  const table = addTableWithHeader(containerSelector, labels);
  const total = new Amount(0, account.commodity);
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
        [shortenAccountName(account.relativeName(p.account), 30), "account"],
        [p.transaction.other(p).account, "account"],
        [p.quantity, "amount"],
        [Amount.clone(total), "amount"],
      ];
      if (options.showLocation) values.push([p.transaction.location, "text"]);
      return values;
    })
    .join("td")
    .classed("amount", ([v, c]) => c == "amount")
    .attr("rowspan", (_, i) => (i == 0 && options.showNotes ? 2 : null))
    .text(([v, c]) => v.toString());
  if (options.showNotes) {
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
