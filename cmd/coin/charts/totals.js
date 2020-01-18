var data = d3.csvParse(d3.select("p#data").text(), d3.autoType)
var svg = d3.select("svg#chart")
var height = +svg.attr("height")
var width = +svg.attr("width")
var x = d3.scaleBand().range([0, width])
var y = d3.scaleLinear().range([height, 0])
var z = d3.scaleOrdinal(d3.schemeCategory10);
var xAxis = d3.axisBottom(x)
var yAxis = d3.axisRight(y)

var layers = d3.stack().keys(data.columns.slice(1))(data)

x.domain(layers[0].map(function(d) { return d.data.Date; }));
y.domain([0, d3.max(layers[layers.length - 1], function(d) { return d[1] })]).nice();

var layer = svg.selectAll(".layer")
    .data(layers)
    .enter().append("g")
    .attr("class", "layer")
    .style("fill", function(d, i) { return z(i); });

layer.selectAll("rect")
    .data(function(d) { return d; })
    .enter().append("rect")
      .attr("x", function(d) { return x(d.data.Date); })
      .attr("y", function(d) { return y(d[0]); })
      .attr("height", function(d) { return y(d[0]) - y(d[1]); })
      .attr("width", x.bandwidth() - 1);

  svg.append("g")
      .attr("class", "axis axis--x")
      .attr("transform", "translate(0," + height + ")")
      .call(xAxis);

  svg.append("g")
      .attr("class", "axis axis--y")
      .attr("transform", "translate(" + width + ",0)")
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