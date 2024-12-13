import {
  Aggregation,
  State,
  addAggregateInput,
  addSubAccountMaxInput,
  emptyElement,
  MainView,
} from "./views";
import { groupWithSubAccounts } from "./utils";
import { Account } from "./account";
import { axisLeft, axisTop } from "d3-axis";
import { scaleLinear, scaleOrdinal, scaleTime } from "d3-scale";
import { schemeCategory10 } from "d3-scale-chromatic";
import { select } from "d3-selection";

export function viewChart(options?: {
  negated?: boolean; // is this negatively denominated account (e.g. Income/Liability)
}) {
  const containerSelector = MainView;
  const account = State.SelectedAccount;
  const opts = { negated: false }; // defaults
  Object.assign(opts, options);
  // clear out the container
  emptyElement(containerSelector);
  addAggregateInput(containerSelector, {
    includeNone: false,
  });
  addSubAccountMaxInput(containerSelector);

  const groupKey = Aggregation[State.View.Aggregate] as d3.TimeInterval;
  const dates = groupKey.range(State.StartDate, State.EndDate);
  const maxAccounts = State.View.AggregatedSubAccountMax;
  const accountGroups = groupWithSubAccounts(account, groupKey, maxAccounts, {
    negated: opts.negated,
  });
  const labelFromAccount = (a: Account | undefined) =>
    a ? account.relativeName(a) : "Other";
  const labels = accountGroups.map((gs) => labelFromAccount(gs.account));
  // compute offsets for each group left to right
  // and max width for the x domain
  let max = 0;
  dates.forEach((_, i) => {
    let offset = 0;
    accountGroups.forEach((gs) => {
      const group = gs.groups[i];
      group.offset = offset;
      let sum = Math.trunc(group.sum.toNumber());
      if (opts.negated) sum = -sum;
      group.width = sum < 0 ? 0 : sum;
      offset += group.width;
    });
    max = max < offset ? offset : max;
  });

  const rowHeight = 15,
    margin = { top: 3 * rowHeight, right: 50, bottom: 50, left: 100 },
    height = dates.length * rowHeight + margin.top + margin.bottom,
    textOffset = (rowHeight * 3) / 4;

  const svg = select(containerSelector)
    .append("svg")
    .attr("id", "chart")
    .attr("width", "100%")
    .attr("height", height + margin.top + margin.bottom);

  let width =
    Math.max(Math.floor(svg.property("width")?.baseVal.value), 800) -
    margin.left -
    margin.right;

  var chart = svg
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  var x = scaleLinear([0, max], [0, width]).nice();
  var y = scaleTime([State.StartDate, State.EndDate], [0, height]);
  var z = scaleOrdinal([0, maxAccounts], schemeCategory10);
  var xAxis = axisTop(x);
  var yAxis = axisLeft(y).ticks(groupKey, "%Y/%m/%d");

  // bar layers
  var layer = chart
    .selectAll(".layer")
    .data(accountGroups)
    .enter()
    .append("g")
    .attr("class", "layer")
    .style("fill", (d, i) => z(i));

  // bars
  layer
    .selectAll("rect")
    .data((d) => d.groups)
    .enter()
    .append("rect")
    .attr("y", (d) => y(d.date))
    .attr("x", (d) => x(d.offset ?? 0))
    .attr("width", (d) => x(d.width ?? 0))
    .attr("height", rowHeight - 1)
    .on("click", (e, d) => console.log(e, d));

  // bar text
  layer
    .selectAll("text")
    .data((d) => d.groups)
    .enter()
    .append("text")
    .text((d) => {
      const v = d.width ?? 0;
      const w = (Math.log10(v) + 1) * 8;
      return v > 0 && x(v) > w ? v : "";
    })
    .attr("x", (d) => x((d.offset ?? 0) + (d.width ?? 0)) - 4)
    .attr("y", (d) => y(d.date) + textOffset);

  // axis
  chart.append("g").attr("class", "axis axis--x").call(xAxis);
  chart.append("g").attr("class", "axis axis--y").call(yAxis);

  var legend = svg
    .selectAll(".legend")
    .data(labels)
    .enter()
    .append("g")
    .attr("class", "legend")
    .attr("transform", "translate(" + margin.left + ",0)");

  var w = x.range()[1] / labels.length;

  legend
    .append("rect")
    .attr("x", (d, i) => w * i)
    .attr("y", 0)
    .attr("width", w)
    .attr("height", rowHeight - 1)
    .style("fill", (d, i) => z(i));

  legend
    .append("text")
    .text((d) => d)
    .attr("x", (d, i) => w * i + 10)
    .attr("y", textOffset);
}
