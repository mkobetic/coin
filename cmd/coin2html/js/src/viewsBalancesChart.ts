import { stratify, treemap } from "d3-hierarchy";
import { select } from "d3-selection";
import { group } from "d3-array";
import {
  addBalanceDepthInput,
  emptyElement,
  MainView,
  State,
  updateAccount,
} from "./views";
import { scaleSequential } from "d3-scale";
import { interpolateBlues } from "d3-scale-chromatic";
import { AccountBalanceAndTotal } from "./utils";

function max(a: number, b: number) {
  return a > b ? a : b;
}

// based on https://observablehq.com/@d3/nested-treemap
export function viewBalancesChart(options?: {
  negated?: boolean; // is this negatively denominated account (e.g. Income/Liability)
}) {
  const containerSelector = MainView;
  const account = State.SelectedAccount;
  const opts = { negated: false }; // defaults
  Object.assign(opts, options);
  emptyElement(containerSelector);
  addBalanceDepthInput(containerSelector);
  const date = State.EndDate;
  // build the hierarchical data structure that d3 treemap expects
  let root = stratify<AccountBalanceAndTotal>()
    .id(({ account }) => account.fullName)
    .parentId(({ account }) =>
      account.parent && account.parent != State.SelectedAccount.parent
        ? account.parent.fullName
        : undefined
    )(account.withAllChildBalances(date));
  // compute the individual node.value that drives the treemap layout
  root = root.sum((a) =>
    max(
      (opts.negated ? -1 : 1) *
        account.commodity.convert(a.balance, date).toNumber(),
      0
    )
  );

  const [width, height] = [1200, 800];
  const tm = treemap<AccountBalanceAndTotal>()
    .size([width, height])
    .padding(4)
    .paddingTop(20);
  const nodes = tm(root);
  const nodesByDepth = Array.from(group(nodes, (d) => d.depth))
    .sort((a, b) => a[0] - b[0])
    .map((d) => d[1])
    .slice(0, State.View.BalanceDepth);

  let uidCounter = 0;

  const svg = select(containerSelector)
    .append("svg")
    .attr("id", "chart")
    .attr("width", "100%")
    .attr("height", height);
  // .attr("viewBox", [0, 0, width, height]);
  // .attr(
  //   "style",
  //   "max-width: 100%; height: auto; overflow: visible; font: 10px sans-serif;"
  // );

  const color = scaleSequential(
    [0, nodesByDepth.length * 1.5],
    interpolateBlues
  );

  const node = svg
    .selectAll("g")
    .data(nodesByDepth)
    .join("g")
    .selectAll("g")
    .data((d) => d)
    .join("g")
    .attr("transform", (d) => `translate(${d.x0},${d.y0})`);

  node
    .append("title")
    .text(({ data }) =>
      data.account.children.length > 0 && !data.balance.isZero
        ? `${data.account.fullName} ${data.total} [ ${data.balance} ]`
        : `${data.account.fullName} ${data.total}`
    );

  node
    .append("rect")
    .attr("id", (d: any) => (d.nodeUid = `node-${uidCounter++}`))
    .attr("fill", (d) => color(d.depth))
    .attr("width", (d) => d.x1 - d.x0)
    .attr("height", (d) => d.y1 - d.y0)
    .on("click", (e, { data }) => {
      State.SelectedAccount = data.account;
      updateAccount();
    });

  // add a clippath to clip the text to the rectangle
  node
    .append("clipPath")
    .attr("id", (d: any) => (d.clipUid = `clip-${uidCounter++}`))
    .append("use")
    .attr("xlink:href", (d: any) => `#${d.nodeUid}`);

  node
    .append("text")
    .attr("clip-path", (d: any) => `url(#${d.clipUid})`)
    .on("click", (e, { data }) => {
      State.SelectedAccount = data.account;
      updateAccount();
    })
    .selectAll("tspan")
    .data(({ data }) => {
      const bits = [data.account.name, data.total.toString()];
      if (data.account.children.length > 0 && !data.balance.isZero)
        bits.push(data.balance.toString());
      return bits;
    })
    .join("tspan")
    .text((d) => d);

  const narrowBoxLimit = 150;
  // if the box is wide, put the spans on the same line
  node
    .filter((d: any) => d.x1 - d.x0 > narrowBoxLimit)
    .selectAll("tspan")
    .attr("dx", 10)
    .attr("y", 15);
  // if the box is narrow, put the spans on separate lines
  node
    .filter((d) => d.x1 - d.x0 <= narrowBoxLimit)
    .selectAll("tspan")
    .attr("x", 10)
    .attr("dy", 15);
}
