var width = d3.select("#map-container").node().getBoundingClientRect().width,
    height = d3.select("#map-container").node().getBoundingClientRect().height;

console.log(width, height);
var svg = d3.select("#map-container").append("svg")
    .attr("width", width)
    .attr("height", height)
    .append("g");

var projection = d3.geoMercator()
    .scale(width / 2 / Math.PI) //http://bl.ocks.org/almccon/fe445f1d6b177fd0946800a48aa59c71
    .translate([width / 2, height / 2]);

var path = d3.geoPath()
    .projection(projection);

var zoom = d3.zoom()
    .scaleExtent([1, 8])
    .on("zoom", zoomed);

d3.select("button")
    .on("click", resetted);


svg.call(zoom);

function zoomed() {
    svg.selectAll("g").attr("transform", d3.event.transform);
}

function resetted() {
    svg.transition()
        .duration(750)
        .call(zoom.transform, d3.zoomIdentity);
}

function gotMap(error, world) {
    if (error) throw error;

    svg.append("g")
        .attr("class", "land")
        .selectAll("path")
        .data(topojson.feature(world, world.objects.land).features)
        .enter().append("path")
        .attr("d", path);

    // svg.append("g")
    //     .attr("class", "boundary")
    //     .selectAll("boundary")
    //     .data([topojson.feature(world, world.objects.countries)])
    //     .enter().append("path")
    //     .attr("d", path);


    //it's ugly to put this here but until i janker around wif $.Deferred()
    //that's what i'm goin to do
    getPoints();
}

function getMap() {
    return d3.json("https://d3js.org/world-50m.v1.json", gotMap);
}

function getPoints(eps) {
    if (typeof(eps) === "undefined") {
        eps = 0.001
    };
    var u = "/v1?" + encodeURIComponent("epsilon=" + eps);
    return d3.json(u, gotPoints);
}

function gotPoints(err, points) {
    if (err) throw err;
    console.log("received " + points.length + " points from ajaxery");

    console.log("the first ten look like dis ", points.slice(0, 10));
    console.log("projected:");
    for (i in points.slice(0, 10)) {
        var p = points[i];
        var c = [p.lat, p.long];
        console.log(projection(c));
    }

    // // add circles to svg //http://bl.ocks.org/phil-pedruco/7745589
    svg.append("g")
        .selectAll("circle")
        .data(points)
        .enter()
        .append("circle")

    // // ?? this puts in wrong spot //...
    //     .attr("cx", function (d) {
    //       var c = [d.lat, d.long];
    //       return projection(c)[0];
    //     })
    //     .attr("cy", function (d) {
    //       var c = [d.lat, d.long];
    //       return projection(c)[1];
    //     })


    // http://gis.stackexchange.com/questions/34769/how-can-i-render-latitude-longitude-coordinates-on-a-map-with-d3
    // ooohhhhh lng,lat, not backbackwards ?
    .attr("transform", function(d) {
        return "translate(" + projection([d.long, d.lat]) + ")";
    })

    .attr("r", "1px")
        .attr("fill", "red");

    // create a first guess for the projection
  var center = d3.geoCentroid(points);
    var scale = 150;
    var offset = [width / 2, height / 2];
    // using the path determine the bounds of the current map and use 
    // these to determine better values for the scale and translation
    var bounds = path.bounds(points);
    var hscale = scale * width / (bounds[1][0] - bounds[0][0]);
    var vscale = scale * height / (bounds[1][1] - bounds[0][1]);
    var scale = (hscale < vscale) ? hscale : vscale;
    var offset = [width - (bounds[0][0] + bounds[1][0]) / 2,
        height - (bounds[0][1] + bounds[1][1]) / 2
    ];

    // new projection
    projection = d3.geoMercator().center(center)
        .scale(scale).translate(offset);
    path = path.projection(projection);

    // add a rectangle to see the bound of the svg
    svg.append("rect").attr('width', width).attr('height', height)
        .style('stroke', 'black').style('fill', 'none');
}

getMap();

//try to get us states too for more microcosmic
// d3.json("https://d3js.org/us-10m.v1.json", function (err, us) {
//   if (err) throw err;
//   // svg.append("g")
//   //   .attr("class", "boundary")
//   //   .selectAll("boundary")
//   //   .data([topojson.feature(us, us.objects.states)])
//   //   .enter().append("path")
//   //   .attr("d", path);
// });
