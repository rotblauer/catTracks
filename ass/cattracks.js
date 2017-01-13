
$(function () {
console.log("cattracks.js is here")

var currentLocation = {};
var catTrackPoint = {
  name: "jl",
  lat: 0.0,
  lng: 0.0
}
var latInput = $("#myLat"),
    lngInput = $("#myLon"),
    submitBtn = $("#submitposter-btn");

submitBtn.click(postPoint);


// post to /populate with whatever is in the form
function postPoint () {
  console.log("poisting pont");

  // should matchupcatup to trackPoints
  var cataData = {
    "name": $("#myName").val(),
    "lat": parseFloat(latInput.val()),
    "long": parseFloat(lngInput.val()),
    // "time": Date.now(),
    "notes": "json web post"
  };

  var jcatdat = JSON.stringify(cataData);
  console.log("jcatd", jcatdat);

  $.post(
    "/populate/",
    jcatdat,
    handlePostSuccess
    );

}

  // data is trackpoint obj
  function handlePostSuccess(data, status) {
    alert("REsponser: " + data + "\n" + "status: " + status);
    //but won't show newpoint because it's not availbe thru gornedered
    initMap();
  }

});
