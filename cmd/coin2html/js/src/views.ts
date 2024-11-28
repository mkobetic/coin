import * as d3 from "d3";
import {
  Account,
  Accounts,
  Aggregation,
  Amount,
  dateToString,
  groupBy,
  groupWithSubAccounts,
  loadEverything,
  MaxDate,
  MinDate,
  Posting,
  Roots,
  State,
  trimToDateRange,
} from "./models";

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
  // Need to load before initializing the UI state.
  loadEverything();
  State.SelectedAccount = Accounts.Assets;
  State.SelectedView = Object.keys(Views.Assets)[0];
  State.StartDate = MinDate;
  State.EndDate = MaxDate;

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
