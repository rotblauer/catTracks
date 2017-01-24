console.log("feeling racy");

var bigPapi = $("#putemhere");


var catStats = `
<div class="row catrow text-center">
</div>
`;

// var timeperiods = ["today", "week", "all"];
var timeperiods = ["today"];
function gotData(data) {
  console.log(data);
  for (i in timeperiods) {
    var periodData = data[timeperiods[i]];
    for (j in periodData["cat"]) {
      var perCat = periodData["cat"][j];

      var catRow = $(catStats);
      bigPapi.append(catRow);

      // var putem = catRow.first(".cat" + timeperiods[i]);

      var speeder = $("<h1>" + j + " - " + perCat.speed_stats.max + "m/s</h1>");
      catRow.append(speeder);
    }
  }

}




function gotErr(err) {
  alert(err);
}

$.getJSON("/api/race", gotData, gotErr);
