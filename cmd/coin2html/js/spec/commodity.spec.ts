import { Commodity } from "../src/models";

test("create commodity", () =>
  expect(new Commodity("CAD", "Canadian Dollar", 2)).toBeTruthy());
