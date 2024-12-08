import { Commodity } from "../src/commodity";

test("create commodity", () =>
  expect(new Commodity("CAD", "Canadian Dollar", 2, "")).toBeTruthy());
