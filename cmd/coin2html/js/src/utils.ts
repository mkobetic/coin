import { Account, Posting } from "./account";
import { Amount, Commodity } from "./commodity";
import { State } from "./views";

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
