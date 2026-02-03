import { select } from "d3-selection";
import { timeMonth, timeWeek, timeYear } from "d3-time";
import {
  renderPostings,
  renderPostingsWithSubAccounts,
  viewRegister,
} from "./viewsRegister";
import { viewAggregatedRegisterChart } from "./viewsAggregatedRegisterChart";
import { Account } from "./account";
import { PostingGroup, shortenAccountName, topN } from "./utils";
import { viewBalances } from "./viewsBalances";
import { viewBalancesChart } from "./viewsBalancesChart";

export const Aggregation = {
  None: null,
  Weekly: timeWeek,
  Monthly: timeMonth,
  Quarterly: timeMonth.every(3),
  Yearly: timeYear,
};

export enum AggregationStyle {
  Flows = "Flows", // sum of flows for the period
  Balances = "Balances", // balance at the end of the period
}

// UI State
export const State = {
  // All these must be set after loading of data is finished, see initializeUI()
  SelectedAccount: undefined as unknown as Account, // currently viewed account
  AccountListRoot: undefined as unknown as Account, // account used to generate the account list
  SelectedView: "Balance",
  StartDate: new Date(),
  EndDate: new Date(),
  ShowClosedAccounts: false,
  View: {
    // Should we recurse into subaccounts
    ShowSubAccounts: false,
    ExcludeSubAccounts: [] as Account[],
    ShowNotes: false, // Show notes in register view
    Aggregate: "None" as keyof typeof Aggregation,
    // How many largest subaccounts to show when aggregating.
    AggregatedSubAccountMax: 5,
    AggregationStyle: AggregationStyle.Flows as AggregationStyle,
    ShowLocation: false, // Show transaction location info
    BalanceDepth: 3, // How many levels of subaccounts to show in balance view
  },
};

// Available view types by account category.
// These define what is offered in the view drop-down.
// All types have Register.
export const Views = {
  Assets: {
    Balances: viewBalances,
    Register: viewRegister,
    "Balances - Chart": viewBalancesChart,
    "Aggregated Register - Chart": viewAggregatedRegisterChart,
  },
  Liabilities: {
    Balances: viewBalances,
    Register: () => viewRegister({ negated: true }),
    "Balances - Chart": () => viewBalancesChart({ negated: true }),
    "Aggregated Register - Chart": () =>
      viewAggregatedRegisterChart({ negated: true }),
  },
  Income: {
    Balances: viewBalances,
    Register: () =>
      viewRegister({
        negated: true,
        aggregatedTotal: true,
      }),
    "Balances - Chart": () => viewBalancesChart({ negated: true }),
    "Aggregated Register - Chart": () =>
      viewAggregatedRegisterChart({ negated: true }),
  },
  Expenses: {
    Balances: viewBalances,
    Register: () =>
      viewRegister({
        aggregatedTotal: true,
      }),
    "Balances - Chart": viewBalancesChart,
    "Aggregated Register - Chart": viewAggregatedRegisterChart,
  },
  Equity: {
    Balances: viewBalances,
    Register: viewRegister,
  },
  Unbalanced: {
    Balances: viewBalances,
    Register: viewRegister,
  },
};

// View components

export function addIncludeSubAccountsInput(containerSelector: string) {
  const container = select(containerSelector);
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

export function addIncludeNotesInput(containerSelector: string) {
  const container = select(containerSelector);
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

export function addShowLocationInput(containerSelector: string) {
  const container = select(containerSelector);
  container
    .append("label")
    .property("for", "showLocation")
    .text("Show Location");
  container
    .append("input")
    .on("change", (e, d) => {
      const input = e.currentTarget as HTMLInputElement;
      State.View.ShowLocation = input.checked;
      updateView();
    })
    .attr("id", "showLocation")
    .attr("type", "checkbox")
    .property("checked", State.View.ShowLocation);
}

export function addSubAccountMaxInput(containerSelector: string) {
  const container = select(containerSelector);
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

export function addBalanceDepthInput(containerSelector: string) {
  const container = select(containerSelector);
  container.append("label").property("for", "balanceDepth").text("Depth");
  container
    .append("input")
    .on("change", (e, d) => {
      const input = e.currentTarget as HTMLInputElement;
      State.View.BalanceDepth = parseInt(input.value);
      updateView();
    })
    .attr("id", "balanceDepth")
    .attr("type", "number")
    .property("value", State.View.BalanceDepth);
}

export function addAggregateInput(
  containerSelector: string,
  options?: {
    includeNone?: boolean;
  },
) {
  const opts = { includeNone: true }; // defaults
  Object.assign(opts, options);
  const container = select(containerSelector);
  container.append("label").property("for", "aggregate").text("Aggregate");
  const aggregate = container.append("select").attr("id", "aggregate");
  aggregate.on("change", (e, d) => {
    const select = e.currentTarget as HTMLSelectElement;
    const selected = select.options[select.selectedIndex].value;
    State.View.Aggregate = selected as keyof typeof Aggregation;
    updateView();
  });
  let data = Object.keys(Aggregation).filter(
    (k) => opts.includeNone || k != "None",
  );
  if (!opts.includeNone && State.View.Aggregate == "None") {
    State.View.Aggregate = data[0] as keyof typeof Aggregation;
    updateAggregationForTimeRange();
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

export function addAggregationStyleInput(containerSelector: string) {
  const container = select(containerSelector);
  const aggregate = container.append("select").attr("id", "aggregationStyle");
  aggregate.on("change", (e, d) => {
    const select = e.currentTarget as HTMLSelectElement;
    const selected = select.options[select.selectedIndex].value;
    State.View.AggregationStyle = selected as AggregationStyle;
    updateView();
  });
  aggregate
    .selectAll("option")
    .data(Object.keys(AggregationStyle))
    .join("option")
    .property("selected", (v) => v == State.View.AggregationStyle)
    .property("value", (v) => v)
    .text((v) => v);
}

export function addExcludedSubAccountsSpan(
  containerSelector: string,
  account: Account,
) {
  const container = select(containerSelector);
  container
    .append("label")
    .property("for", "excludedSubAccounts")
    .text("Exclude");
  const aggregate = container.append("span").attr("id", "excludedSubAccounts");
  aggregate
    .selectAll("span")
    .data(State.View.ExcludeSubAccounts)
    .join("span")
    .on("click", (e, d) => {
      var i = State.View.ExcludeSubAccounts.indexOf(d);
      State.View.ExcludeSubAccounts.splice(i, 1);
      updateView();
    })
    .text((v) => ` ${account.relativeName(v)}`);
}

// UI Node Selectors

export const RootAccountSelect = "#sidebar select#root";
export const AccountList = "#sidebar ul#accounts";

export const ViewSelect = "#main #controls select#view";
export const StartDateInput = "#main #controls input#start";
export const EndDateInput = "#main #controls input#end";
export const ShowClosedAccounts = "#main #controls input#closedAccounts";
export const AccountName = "#main output#account span#name";
export const AccountCommodity = "#main output#account span#commodity";
export const MainView = "#main section#view";
export const Details = "div#details";

export function emptyElement(selector: string) {
  (select(selector).node() as Element).replaceChildren();
}

// UI Events

export function updateView() {
  const account = State.SelectedAccount.getRootAccount();
  const selectedViews = Views[account.name as keyof typeof Views];
  const view = selectedViews[State.SelectedView as keyof typeof selectedViews];
  view();
}

export function updateAggregationForTimeRange() {
  if (State.View.Aggregate == "None") return;
  const days =
    (State.EndDate.getTime() - State.StartDate.getTime()) /
    (1000 * 60 * 60 * 24);
  if (days < 180) State.View.Aggregate = "Weekly";
  else if (days < 3 * 180) State.View.Aggregate = "Monthly";
  else if (days < 5 * 365) State.View.Aggregate = "Quarterly";
  else State.View.Aggregate = "Yearly";
}

export function updateAccount() {
  const account = State.SelectedAccount;
  const spans = select(AccountName)
    .selectAll("span.account")
    .data(account.withAllParents())
    .join("span")
    .classed("account", true)
    .text((d) => (d.parent ? ":" : ""));
  spans
    .append("a")
    .text((acc: Account) => acc.name)
    .on("click", (e: Event, acc: Account) => {
      State.SelectedAccount = acc;
      if (acc.isParentOf(State.AccountListRoot)) updateAccounts();
      else updateAccount();
    });
  select(AccountCommodity).text(` (${account.commodity})`);
  updateView();
}

export function addViewSelect() {
  emptyElement(ViewSelect);
  const account = State.SelectedAccount.getRootAccount();
  const selectedViews = Object.keys(Views[account.name as keyof typeof Views]);
  if (!selectedViews.includes(State.SelectedView))
    State.SelectedView = selectedViews[0];
  select(ViewSelect)
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
export function addAccountList() {
  State.AccountListRoot = State.SelectedAccount;
  select(AccountList)
    .selectAll("li")
    .data(State.AccountListRoot.allChildren())
    .join("li")
    .text((d) => shortenAccountName(State.SelectedAccount.relativeName(d), 40))
    .on("click", (e: Event) => {
      State.SelectedAccount = (e.currentTarget as liWithAccount).__data__;
      updateAccount();
    })
    .on("dblclick", (e: Event) => {
      State.SelectedAccount = (e.currentTarget as liWithAccount).__data__;
      updateAccounts();
    });
}

export function updateAccounts() {
  addViewSelect();
  addAccountList();
  updateAccount();
}

export function showDetails(g: PostingGroup, withSubaccounts = false) {
  emptyElement(Details);
  const details = select(Details);
  details
    .insert("a")
    .text("X")
    .on("click", () => details.attr("hidden", true));
  const account = State.SelectedAccount;
  const data = topN(g.postings, 20, account.commodity);
  const options = {
    negated: false,
    showLocation: true,
  };
  if (withSubaccounts)
    renderPostingsWithSubAccounts(account, data, Details, options);
  else renderPostings(account, data, Details, options);

  details.attr("hidden", null);
}
export function addTableWithHeader(
  containerSelector: string,
  labels: string[],
) {
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
