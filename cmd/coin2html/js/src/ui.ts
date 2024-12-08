import * as d3 from "d3";
import { viewRegister } from "./register";
import { viewChart } from "./chart";
import {
  Account,
  Accounts,
  loadAccounts,
  MaxDate,
  MinDate,
  Roots,
} from "./account";
import { dateToString } from "./utils";
import { loadCommodities } from "./commodity";

export const Aggregation = {
  None: null,
  Weekly: d3.timeWeek,
  Monthly: d3.timeMonth,
  Quarterly: d3.timeMonth.every(3),
  Yearly: d3.timeYear,
};

// UI State
export const State = {
  // All these need to be set again after loading of date is finished.
  SelectedAccount: Accounts.Assets,
  SelectedView: "Register",
  StartDate: MinDate,
  EndDate: MaxDate,
  ShowClosedAccounts: false,
  View: {
    // Should we recurse into subaccounts
    ShowSubAccounts: false,
    ShowNotes: false, // Show notes in register view
    Aggregate: "None" as keyof typeof Aggregation,
    // How many largest subaccounts to show when aggregating.
    AggregatedSubAccountMax: 5,
  },
};

// View types by account category.
// All types have Register.
export const Views = {
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

export function addIncludeSubAccountsInput(containerSelector: string) {
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

export function addIncludeNotesInput(containerSelector: string) {
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

export function addSubAccountMaxInput(containerSelector: string) {
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

export function addAggregateInput(
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

// UI Node Selectors

export const RootAccountSelect = "#sidebar select#root";
export const AccountList = "#sidebar ul#accounts";

export const ViewSelect = "#main #controls select#view";
export const StartDateInput = "#main #controls input#start";
export const EndDateInput = "#main #controls input#end";
export const ShowClosedAccounts = "#main #controls input#closedAccounts";
export const AccountOutput = "#main output#account";
export const MainView = "#main section#view";

export function emptyElement(selector: string) {
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
  // d3.select(AccountOutput).text(account.fullName);
  const spans = d3
    .select(AccountOutput)
    .selectAll("span")
    .data(account.withAllParents())
    .join("span")
    .text((d) => (d.parent ? ":" : ""));
  spans
    .append("a")
    .text((acc: Account) => acc.name)
    .on("click", (e: Event, acc: Account) => {
      State.SelectedAccount = acc;
      updateAccount();
    });
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
  loadCommodities();
  loadAccounts();
  State.SelectedAccount = Accounts.Assets;
  State.SelectedView = Object.keys(Views.Assets)[0];
  State.StartDate = MinDate;
  State.EndDate = MaxDate;
  State.ShowClosedAccounts = false;

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
  d3.select(ShowClosedAccounts)
    .on("change", (e: Event) => {
      const input = e.currentTarget as HTMLInputElement;
      State.ShowClosedAccounts = input.checked;
      updateAccounts();
    })
    .property("checked", State.ShowClosedAccounts);

  // trigger account selection
  updateAccounts();
}

initializeUI();
