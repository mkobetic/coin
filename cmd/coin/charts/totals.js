var data = d3.csvParseRows(d3.select("p#data").html(),function(d){
    return d.map(function(v,i){
        if (i == 0) {
            var t = d3.timeParse("%Y/%M")(v)
            return t == null ? v : t
        } else {
            var n = +v
            return isNaN(n) ? v : n
        }
    })
})
var header = data[0]
data = data.slice(1)

