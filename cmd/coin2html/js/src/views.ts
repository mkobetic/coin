import { select } from "d3-selection";
import { timeMonth, timeWeek, timeYear } from "d3-time";
import { viewRegister } from "./register";
import { viewChart } from "./chart";
import { Account } from "./account";

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
  SelectedAccount: undefined as unknown as Account,
  SelectedView: "Register",
  StartDate: new Date(),
  EndDate: new Date(),
  ShowClosedAccounts: false,
  View: {
    // Should we recurse into subaccounts
    ShowSubAccounts: false,
    ShowNotes: false, // Show notes in register view
    Aggregate: "None" as keyof typeof Aggregation,
    // How many largest subaccounts to show when aggregating.
    AggregatedSubAccountMax: 5,
    AggregationStyle: AggregationStyle.Flows as AggregationStyle,
    ShowLocation: false, // Show transaction location info
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

export function addAggregateInput(
  containerSelector: string,
  options?: {
    includeNone?: boolean;
  }
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
  (select(selector).node() as Element).replaceChildren();
}

// UI Events

export function updateView() {
  const account = State.SelectedAccount.getRootAccount();
  const selectedViews = Views[account.name as keyof typeof Views];
  const view = selectedViews[State.SelectedView as keyof typeof selectedViews];
  view();
}

export function updateAccount() {
  const account = State.SelectedAccount;
  // d3.select(AccountOutput).text(account.fullName);
  const spans = select(AccountOutput)
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
  const account = State.SelectedAccount;
  select(AccountList)
    .selectAll("li")
    .data(account.allChildren())
    .join("li")
    .text((d) => State.SelectedAccount.relativeName(d))
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
