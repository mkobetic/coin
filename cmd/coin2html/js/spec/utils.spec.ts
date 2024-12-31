import { Account, Posting, Transaction } from "../src/account";
import { Amount, amount, commodity } from "../src/commodity";
import { topN } from "../src/utils";
import { setupCommodities } from "./setup";

setupCommodities();

describe("topN", () => {
  const CAD = commodity`CAD`;
  const t = new Transaction(new Date(), "test");
  const a = new Account("test", "test", CAD);
  test.each([
    [`2 CAD, 5 CAD, 3 CAD, 1 CAD, 4 CAD`, 2, `5.00 CAD, 4.00 CAD`],
    [`2 CAD, -5 CAD, -1 CAD, 4 CAD`, 3, `-5.00 CAD, 4.00 CAD, 2.00 CAD`],
  ])(`%#: %s top %i`, (input, n, expected) => {
    const postings = input.split(", ").map((s) => {
      const amt = Amount.parse(s);
      return new Posting(t, a, amt, amt);
    });
    const top = topN(postings, n, CAD);
    expect(top.map((p) => p.quantity.toString())).toEqual(expected.split(", "));
  });
});
