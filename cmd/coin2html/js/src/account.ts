import { Amount, Commodity } from "./commodity";
import { State } from "./views";
import {
  AccountPostingGroups,
  dateToString,
  groupBy,
  trimToDateRange,
} from "./utils";

/**
 *  Account, Posting and Transaction
 */

export class Account {
  children: Account[] = [];
  postings: Posting[] = [];
  constructor(
    readonly name: string,
    readonly fullName: string,
    readonly commodity: Commodity,
    readonly parent?: Account,
    readonly closed?: Date,
    readonly location?: string
  ) {
    if (parent) {
      parent.children.push(this);
    }
  }
  allChildren(): Account[] {
    return this.children.reduce(
      (total: Account[], acc: Account) =>
        State.ShowClosedAccounts || !acc.closed
          ? total.concat([acc, ...acc.allChildren()])
          : total,
      []
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
  withAllParents(): Account[] {
    return this.parent ? this.parent.withAllParents().concat(this) : [this];
  }
  getRootAccount(): Account {
    return this.parent ? this.parent.getRootAccount() : this;
  }
  isParentOf(a: Account): boolean {
    if (!a.parent) return false;
    if (a.parent == this) return true;
    return this.isParentOf(a.parent);
  }
}

export interface Tags {
  [key: string]: string;
}

export class Posting {
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

export class Transaction {
  postings: Posting[] = [];
  constructor(
    readonly posted: Date,
    readonly description: string,
    readonly notes?: string[],
    readonly code?: string,
    readonly location?: string
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

type importedAccounts = Record<
  string,
  {
    name: string;
    fullName: string;
    commodity: string;
    parent: string;
    closed?: string;
    location: string;
  }
>;
type importedTransactions = {
  posted: string;
  description: string;
  location: string;
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
export let MinDate = new Date();
export let MaxDate = new Date(0);

export const Accounts: Record<string, Account> = {};
export const Roots: Account[] = [];
export function loadAccounts(source: string) {
  const importedAccounts = JSON.parse(source) as importedAccounts;
  for (const impAccount of Object.values(importedAccounts)) {
    if (impAccount.name == "Root") continue;
    const parent = Accounts[impAccount.parent];
    const account = new Account(
      impAccount.name,
      impAccount.fullName,
      Commodity.find(impAccount.commodity),
      parent,
      impAccount.closed ? new Date(impAccount.closed) : undefined,
      impAccount.location
    );
    Accounts[account.fullName] = account;
    if (!parent) {
      Roots.push(account);
    }
  }
}

export function loadTransactions(source: string) {
  const importedTransactions = JSON.parse(source) as importedTransactions;
  for (const impTransaction of Object.values(importedTransactions)) {
    const posted = new Date(impTransaction.posted);
    if (posted < MinDate) MinDate = posted;
    if (MaxDate < posted) MaxDate = posted;
    const transaction = new Transaction(
      posted,
      impTransaction.description,
      impTransaction.notes,
      impTransaction.code,
      impTransaction.location
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
