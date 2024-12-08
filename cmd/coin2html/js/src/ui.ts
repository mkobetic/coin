import * as d3 from "d3";
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
import {
  EndDateInput,
  RootAccountSelect,
  ShowClosedAccounts,
  StartDateInput,
  State,
  updateAccounts,
  updateView,
  Views,
} from "./views";

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
