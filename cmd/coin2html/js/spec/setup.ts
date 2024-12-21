import { Commodities, Commodity } from "../src/commodity";

export function setupCommodities() {
  for (const [id, decimals] of Object.entries({
    USD: 2,
    CAD: 2,
    EUR: 2,
    CZK: 2,
  }))
    if (!Commodities[id]) Commodities[id] = new Commodity(id, id, decimals, "");
}
