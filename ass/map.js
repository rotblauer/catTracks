var width = d3.select("body").node().getBoundingClientRect().width,
    height = d3.select("body").node().getBoundingClientRect().height;

var svg = d3.select("body").append("svg")
    .attr("width", width)
    .attr("height", height)
    .append("g");

var projection = d3.geoMercator()
      .scale(width / 2 / Math.PI) //http://bl.ocks.org/almccon/fe445f1d6b177fd0946800a48aa59c71
      .translate([width / 2, height / 2]);

var path = d3.geoPath()
    .projection(projection);

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
function getMap (){
  return d3.json("https://d3js.org/world-50m.v1.json", gotMap);
}

function getPoints (eps) {
  if (typeof(eps) === "undefined") { eps = 0.001 };
  var u = "/v1?" + encodeURIComponent("epsilon=" + eps);
  return d3.json(u, gotPoints);
}

function gotPoints (err, points) {
  if (err) throw err;
  console.log("received " + points.length + " points from ajaxery");

  console.log("the first ten look like dis ", points.slice(0,10));
  console.log("projected:");
  for (i in points.slice(0,10)) {
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
    .attr("transform", function (d) {
      return "translate(" + projection([d.long,d.lat]) + ")";
     })

    .attr("r", "4px")
    .attr("fill", "red");

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
