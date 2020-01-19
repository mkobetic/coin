var data = d3.csvParse(d3.select("p#data").text(), d3.autoType)

var margin = {top: 30, right: 20, bottom: 20, left: 50},
    width = 800 - margin.left - margin.right,
    height = data.length*30 + margin.top + margin.bottom;

var svg = d3.select("body").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

var x = d3.scaleLinear().range([0, width])
var y = d3.scaleBand().range([0, height])
var z = d3.scaleOrdinal(d3.schemeCategory10);
var xAxis = d3.axisTop(x)
var yAxis = d3.axisLeft(y)

var layers = d3.stack().keys(data.columns.slice(1))(data)

y.domain(layers[0].map(function(d) { return d.data.Date; }));
x.domain([0, d3.max(layers[layers.length - 1], function(d) { return d[1] })]).nice();

var layer = svg.selectAll(".layer")
    .data(layers)
    .enter().append("g")
    .attr("class", "layer")
    .style("fill", function(d, i) { return z(i); });

layer.selectAll("rect")
    .data(function(d) { return d; })
    .enter().append("rect")
      .attr("y", function(d) { return y(d.data.Date); })
      .attr("x", function(d) { return x(d[0]); })
      .attr("width", function(d) { return x(d[1]) - x(d[0]); })
      .attr("height", y.bandwidth() - 1);

  svg.append("g")
      .attr("class", "axis axis--x")
      .call(xAxis);

  svg.append("g")
      .attr("class", "axis axis--y")
      .call(yAxis);


// function(d){
//     return d.map(function(v,i){
//         if (i == 0) {
//             var t = d3.timeParse("%Y/%M")(v)
//             return t == null ? v : t
//         } else {
//             var n = +v
//             return isNaN(n) ? v : n
//         }
//     })
// })