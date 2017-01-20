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

var circles;

d3.select("button")
    .on("click", resetted);

function zoomed() {
    var transform = d3.event.transform;
    svg.selectAll("g").attr("transform", transform);
    // circles.attr("r", "1px");
    // var t = d3.transform(d3.select(this).attr("transform")).translate;//maintain aold marker translate
    // return "translate(" + t[0] +","+ t[1] + ")scale("+1/scale+")";
    //circles.attr("transform", function(d) {
    // var p = projection([d.long, d.lat]);
    // return "translate(" + transform.applyX(p[0]) +
    // "," + transform.applyY(p[1]) + ")";
    // return "translate(" + transform.applyX(d[0]) + "," + transform.applyY(d[1]) + ")";
    // var t = d3.transform(d3.select(this).)
    // });

    // circles.attr("transform", function (d) {
    //   // var t = d3.transform(d3.select(this).attr("transform")).translate;
    //   // var p = projection([transform.applyY( d.long ), transform.applyX( d.lat )]);
    //   // return "translate(" + p[0] + "," + p[1] + ")";
    //   // return "translate(" + transform.applyX(p[1]) + "," + transform.applyY(p[0]) + ")";
    //   // return "translate(" + transform.applyX(t[0]) + "," + transform.applyY(t[1]) + ")";
    //   // return "translate( " + p  +  ")";
    // });

}

function resetted() {
    svg.transition()
        .duration(450)
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
    return d3.json("https://d3js.org/world-110m.v1.json", gotMap); //50m
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
    circles = svg.append("g")
        .selectAll("circle")
        .data(points)
        .enter()
        .append("circle")

    // // http://gis.stackexchange.com/questions/34769/how-can-i-render-latitude-longitude-coordinates-on-a-map-with-d3
    // // ooohhhhh lng,lat, not backbackwards ?
    .attr("transform", function(d) {
            return "translate(" + projection([d.long, d.lat]) + ")";
        })
        // .attr("transform", transform)

    .attr("r", "1px")
        .attr("fill", "red");

    // //http://stackoverflow.com/questions/14492284/center-a-map-in-d3-given-a-geojson-object
    // Compute the bounds of a feature of interest, then derive scale & translate.
    var b = path.bounds(points),
        s = .95 / Math.max((b[1][0] - b[0][0]) / width, (b[1][1] - b[0][1]) / height),
        t = [(width - s * (b[1][0] + b[0][0])) / 2, (height - s * (b[1][1] + b[0][1])) / 2];

    // Update the projection to use computed scale & translate.
    projection
        .scale(s)
        .translate(t);
    // add a rectangle to see the bound of the svg
    svg.append("rect").attr('width', width).attr('height', height)
        .style('fill', 'none').style('pointer-events', 'all')
        .call(zoom);
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
