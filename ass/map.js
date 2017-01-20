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
    .scaleExtent([1, 1000])
    .on("zoom", zoomed);

var circles;

d3.select("button")
    .on("click", resetted);

function zoomed() {
    var transform = d3.event.transform;
    svg.selectAll("g").attr("transform", transform);

    projection.translate(transform)
        .scale(transform);

    circles.attr("r", function(d) {
        return 1 / transform.k;
    });
}

function resetted() {
    svg
        .transition()
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

    svg.append("g")
        .attr("class", "boundary")
        .selectAll("boundary")
        .data([topojson.feature(world, world.objects.countries)])
        .enter().append("path")
        .attr("d", path);

    //it's ugly to put this here but until i janker around wif $.Deferred()
    //that's what i'm goin to do
  getBerlin();
}

function gotBerlin(err, berlin) {
  if (err) throw err;

  svg.append("g")
    .attr("class", "road")
    .selectAll("road")
    .data([topojson.feature(berlin, berlin.objects.tracts)])
    .enter().append("path")
    .attr("d", path);

  //noncallback
  getPoints();
}
function getBerlin() {
  return d3.json("/ass/berlin2/berlin-geo-topo-simple-quant.json", gotBerlin);
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

function drawCircles(points) {
    // // add circles to svg //http://bl.ocks.org/phil-pedruco/7745589
    circles = svg.append("g")
        .selectAll("circle")
        .data(points)
        .enter()
        .append("circle")
        .attr("transform", function(d) {
            return "translate(" + projection([d.long, d.lat]) + ")";
        })
        .attr("fill", function(d) {
            if (d.name === "Big Papa") {
                return "red";
            } else if (d.name === "RyePhone") {
              return "blue";
            }
            return "green";
        })
        .attr("r", 1);
}

function getFitBounds(points) {

    var bounds = path.bounds(points),
        dx = bounds[1][0] - bounds[0][0],
        dy = bounds[1][1] - bounds[0][1],
        x = (bounds[0][0] + bounds[1][0]) / 2,
        y = (bounds[0][1] + bounds[1][1]) / 2,
        scale = .9 / Math.max(dx / width, dy / height),
        translate = [width / 2 - scale * x, height / 2 - scale * y];

    return {
        t: translate,
        s: scale
    };
}

function gotPoints(err, points) {
    if (err) throw err;

    drawCircles(points);

  // var b = getFitBounds(points);
  // svg.selectAll("g").transition()
  //   .duration(750)
  //   .style("stroke-width", 1.5 / b.s + "px")
  //   .attr("transform", "translate(" + b.t + ")scale(" + b.s + ")");

    // add a rectangle to see the bound of the svg
    var rect = svg.append("rect").attr('width', width).attr('height', height)
        .style('fill', 'none').style('pointer-events', 'all')
          .call(zoom);

}

getMap();
