var data = d3.csvParse(d3.select("p#data").text(), d3.autoType)

var rowHeight = 15,
    margin = {top: rowHeight+20, right: 20, bottom: 20, left: 50},
    width = 800 - margin.left - margin.right,
    height = data.length*rowHeight + margin.top + margin.bottom;

var svg = d3.select("body").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)

var chart = svg.append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

var x = d3.scaleLinear().range([0, width])
var y = d3.scaleBand().range([0, height])
var z = d3.scaleOrdinal(d3.schemeCategory10);
var xAxis = d3.axisTop(x)
var yAxis = d3.axisLeft(y)

var layers = d3.stack().keys(data.columns.slice(1))(data)

y.domain(layers[0].map(function(d) { return d.data.Date; }));
x.domain([0, d3.max(layers[layers.length - 1], function(d) { return d[1] })]).nice();

var layer = chart.selectAll(".layer")
    .data(layers)
    .enter().append("g")
    .attr("class", "layer")
    .style("fill", function(d, i) { return z(i); })
    
layer.selectAll("rect")
    .data(function(d) { return d; })
    .enter().append("rect")
    .attr("y", function(d) { return y(d.data.Date); })
    .attr("x", function(d) { return x(d[0]); })
    .attr("width", function(d) { return x(d[1]) - x(d[0]); })
    .attr("height", y.bandwidth() - 1);

layer.selectAll("text")
    .data(function(d) { return d; })
    .enter().append("text")
    .text(function(d){
        v = d[1]-d[0];
        w = (Math.log10(v) + 1) * 8;
        return v > 0 && x(v) > w ? Math.trunc(v) : ""; })
    .attr("x", function(d) { return x(d[1])-2; })
    .attr("y", function(d) { return y(d.data.Date)+y.bandwidth()*3/4; })

chart.append("g")
    .attr("class", "axis axis--x")
    .call(xAxis);

chart.append("g")
    .attr("class", "axis axis--y")
    .call(yAxis);

var legend = svg.selectAll(".legend")
    .data(data.columns.slice(1))
    .enter().append("g")
    .attr("class", "legend")
    .attr("transform", "translate(" + margin.left + ",0)")
    
var w = (x.domain()[1]-10)/(data.columns.length-1)
    
legend.append("rect")
    .attr("x", function(d,i) { return x(w*i); })
    .attr("y", 0)
    .attr("width", function(d) { return w; })
    .attr("height", y.bandwidth() - 1)
    .style("fill", function(d, i) { return z(i); });

legend.append("text")
    .text(function(d) {return d;})
    .attr("x",function(d,i) {return x(w*i+10);})
    .attr("y", y.bandwidth()*3/4);


