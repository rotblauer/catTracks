
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

  $('#getgeolocation-btn').click(getGeo);

// post to /populate with whatever is in the form
function postPoint () {
  console.log("poisting pont");

  // should matchupcatup to trackPoints
  var cataData = {
    "name": $("#myName").val(),
    "lat": parseFloat(latInput.val()),
    "long": parseFloat(lngInput.val()),
    "notes": "json web post"
  };

  var jcatdat = JSON.stringify(cataData);

  $.post(
    "/populate/",
    jcatdat,
    handlePostSuccess
    );
}

  // data is trackpoint obj
  function handlePostSuccess(data, status) {
    alert("REsponser: " + data + "\n" + "status: " + status);

    // addPointMarker(map, 0, JSON.parse( data )); //can't get this to work.

    //workaround for adding a point to the map wthout requerying the swerver
    var points = $("#trackPointsData").text();
    points = JSON.parse(points);
    data = JSON.parse( data ); //swerver return succesffuly created point
    if (points !== null) {
      points.splice(0, 0, data); //put new point in i=0
    } else {
      points = [];
      points.push(data);
    }
    $("#trackPointsData").text(JSON.stringify( points )); //and update our go-->js gobetween
    initMap();
  }

});

var bounds;
var positions = [];
var map;

function initMap() {
  positions = [];
  bounds = new google.maps.LatLngBounds();
  map = new google.maps.Map(document.getElementById('map'), {
    zoom: 3,
    center: {lat: 38.6270, lng: -90.1994},
    mapTypeId: 'terrain'
  });
  addTrackPointsToMap(map);
}

function addTrackPointsToMap(map) {
  var pointsData = JSON.parse($("#trackPointsData").text());
  // console.log("JSON parsed .TrackPoints:", pointsData);
  if (Array.isArray(pointsData)) {
    for (var i = 0; i < pointsData.length; i++) {
      addPointMarker(map, i, pointsData[i]);
      console.log("marker: ", pointsData[i]);
    }
  }
  drawFlightPath(map);
  map.fitBounds(bounds);
}

function addPointMarker (map, index, trackPoint) {
  console.log("adding point marker for map");
  var infoWindow = new google.maps.InfoWindow(), marker;
  var position = new google.maps.LatLng(trackPoint.lat, trackPoint.long);
  bounds.extend(position);
  positions.push(position);
  var markerObj = {
    position: position,
    map: map,
    title: trackPoint.name
  };

  if (index == 0) markerObj = $.extend({}, markerObj, {animation: google.maps.Animation.BOUNCE});
  var marker = new google.maps.Marker(markerObj);

  //https://wrightshq.com/playground/placing-multiple-markers-on-a-google-map-using-api-3/
  google.maps.event.addListener(marker, 'click', (function(marker, index) {
    return function() {
      infoContent = "<h3>On this day</h3>, " + trackPoint.time + ", the cat was running " + trackPoint.speed + " mph at an elevation of" + trackPoint.elevation + " ft";
      infoWindow.setContent(infoContent);
      infoWindow.open(map, marker);
      map.setCenter(marker.getPosition());
    }
  })(marker, index));
}

function drawFlightPath(map) {
  var flightPath = new google.maps.Polyline({
    path: positions,
    strokeColor: "#0000FF",
    strokeOpacity: 0.3,
    strokeWeight: 3
  });
  flightPath.setMap(map);
}
