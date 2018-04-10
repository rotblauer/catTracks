console.log("feeling racy");

var bigPapi = $("#putemhere");


var catStats = `
<div class="row catrow ">
<div class="col-sm-4 catcol all">
<small class="text-muted">Alltime</small>
</div>
<div class="col-sm-4 catcol week">
<small class="text-muted">This week</small>
</div>
<div class="col-sm-4 catcol today">
<small class="text-muted">Today</small>
</div>
</div>
`;


// var timeperiods = ["today", "week", "all"];
var timeperiods = ["today"];

function getUniqueCats(data) {
  var cats = [];
  for (i in data) {
    var timeperioddata = data[i];
    for (i in timeperioddata.cat) {
      var name = i;
      if (cats.indexOf(i) < 0) {
        cats.push(i);
      }
    }
  }
  // console.log("got cats", cats);
  return cats;
}


//TODO avg speed is borken on swerver side trackpoints/stats.go
function makeSpeedometer(id, maxSpeed, avgSpeed, milage) {
  console.log("speedo", id);
  console.log("maxspeed", maxSpeed, "avgSpeed", avgSpeed);
  // console.log($("#"+id).length);
   var svg = d3.select("#" + id)
                .append("svg:svg")
                .attr("width", 400)
                .attr("height", 400);

  // //uggggly
  // var avgSpeedGauge = iopctrl.arcslider()
  //       .radius(120)
  //       .events(false)
  //       .indicator(iopctrl.defaultGaugeIndicator);
  // avgSpeedGauge.axis().orient("in")
  //   .normalize(true)
  //   .ticks(12)
  //   .tickSubdivide(3)
  //   .tickSize(10, 8, 10)
  //   .tickPadding(5)
  //   .scale(d3.scale.linear()
  //          .domain([0, 30])
  //          .range([-3*Math.PI/4, 3*Math.PI/4]));


        var maxSpeedGauge = iopctrl.arcslider()
                .radius(120)
                .events(false)
                .indicator(iopctrl.defaultGaugeIndicator);
        maxSpeedGauge.axis().orient("in")
                .normalize(true)
                .ticks(12)
                .tickSubdivide(3)
                .tickSize(10, 8, 10)
                .tickPadding(5)
                .scale(d3.scale.linear()
                        .domain([0, 30])
                        .range([-3*Math.PI/4, 3*Math.PI/4]));


        var segDisplay = iopctrl.segdisplay()
                .width(80)
                .digitCount(6)
                .negative(false)
                .decimals(0);

        svg.append("g")
                .attr("class", "segdisplay")
                .attr("transform", "translate(130, 200)")
                .call(segDisplay);

        svg.append("g")
                .attr("class", "gauge max")
                .call(maxSpeedGauge);

  // svg.append("g")
  //   .attr("class", "gauge avg")
  //   .call(avgSpeedGauge);


        segDisplay.value(milage);
  // avgSpeedGauge.value(avgSpeed);
  maxSpeedGauge.value(maxSpeed);
}


function gotData(data) {
  console.log(data);
  var cats = getUniqueCats(data);

  for (i in cats) {
    var cat = cats[i];

    if (cat.length === 0) { continue; }

    var catrow = $(catStats);
    bigPapi.append(catrow);

    for (i in data) {

      var col = catrow.children("." + i).first();
      col.attr("id", ( cat+i ).replace(' ', ""));



      console.log("i", i);
      console.log("datai", data[i]);
      var tp = data[i]["cat"][cat];


      if (typeof(tp) === "undefined") { continue; }
      var count = tp.count;
      var counter = $("<p>" + count + " points</p>");
      catrow.children("." + i).first().append(counter);

      var distance = Math.floor( tp.distance );
      // var distancer = $("<p>" + distance + "m</p>");
      // catrow.children("." + i).first().append(distancer);


      makeSpeedometer(( cat+i ).replace(' ', ""), Math.floor( tp.speed_stats.max ), Math.floor(tp.speed_stats.mean),distance);



    }

    catrow.prepend("<div class='col-sm-12 '><small >" + cat + "</small></div>");

  }


  // for (i in timeperiods) {
  //   var periodData = data[timeperiods[i]];
  //   for (j in periodData["cat"]) {
  //     var perCat = periodData["cat"][j];
  //     var catRow = $(catStats);
  //     bigPapi.append(catRow);
  //     // var putem = catRow.first(".cat" + timeperiods[i]);
  //     var speeder = $("<h1>" + j + " - " + perCat.speed_stats.max + "m/s</h1>");
  //     catRow.append(speeder);
  //   }
  // }

}




function gotErr(err) {
  alert(err);
}

$.getJSON("/api/race", gotData, gotErr);
