import {
  addBalanceDepthInput,
  addTableWithHeader,
  emptyElement,
  MainView,
  State,
} from "./views";

export function viewBalances(options?: {
  negated?: boolean; // is this negatively denominated account (e.g. Income/Liability)
}) {
  const containerSelector = MainView;
  const account = State.SelectedAccount;
  const opts = { negated: false, aggregatedTotal: false };
  Object.assign(opts, options);
  // clear out the container
  emptyElement(containerSelector);
  addBalanceDepthInput(containerSelector);
  const labels = ["Balance", "Total", "Account"];
  const table = addTableWithHeader(containerSelector, labels);
  let balances = account.withAllChildBalances(State.EndDate);
  if (!State.ShowClosedAccounts) {
    balances = balances.filter((b) => !b.account.isClosed(State.EndDate));
  }
  balances = balances.filter(
    (b) => b.account.depthFrom(account) <= State.View.BalanceDepth
  );

  const rows = table.append("tbody").selectAll("tr").data(balances).enter();
  rows
    .append("tr")
    .classed("even", (_, i) => i % 2 == 0)
    .selectAll("td")
    .data((b) => [
      [b.balance, "amount"],
      [b.total, "amount"],
      [account.relativeName(b.account), "account"],
    ])
    .join("td")
    .classed("amount", ([v, c]) => c == "amount")
    .text(([v, c]) => v.toString());
}
