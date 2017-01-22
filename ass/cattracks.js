
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

  function getRandomDev(i) {
    return Math.random()*parseFloat(i+1);
  }
// post to /populate with whatever is in the form
function postPoint () {
  console.log("poisting pontificus");

  // should matchupcatup to trackPoints
  var cataData = {
    "name": $("#myName").val(),
    "lat": parseFloat(latInput.val()),
    "long": parseFloat(lngInput.val()),
    "notes": "json web post"
  };

  //one for all
  var catdatas = [];
  for (var i = 0; i < 3; i++) {
    var n = cataData;
    var r = getRandomDev(i);//randomnotrandom
    n.lat = cataData.lat + r;
    n.long = cataData.long + r;
    catdatas.push(n);
  }

  var jcatdat = JSON.stringify(catdatas);
  console.log(jcatdat);

  $.post(
    "/populate/",
    jcatdat,
    handlePostSuccess
    );
}

  // data is trackpoint obj

    // I think I broke this
  function handlePostSuccess(data, status) {
    alert("REsponser: " + data + "\n" + "status: " + status);

    // addPointMarker(map, 0, JSON.parse( data )); //can't get this to work.

    //workaround for adding a point to the map wthout requerying the swerver
    var points = $("#trackPointsData").text();
    points = JSON.parse(points);
    data = JSON.parse( data ); //swerver return succesffuly created point
    if (points !== null) {
      for (i in data) {
        points.splice(0, 0, data[i]); //put new point in i=0
      }
    } else {
      points = [];
      for (i in data) {
        points.push(data[i]);
      }
    }
    $("#trackPointsData").text(JSON.stringify( points )); //and update our go-->js gobetween
    initMap();
  }

  console.log("RAGNER RAGNER -- ", $("input#daRslider").length);
  $('input#daRslider').rangeslider({

    // Feature detection the default is `true`.
    // Set this to `false` if you want to use
    // the polyfill also in Browsers which support
    // the native <input type="range"> element.
    polyfill: true,

    // Default CSS classes
    rangeClass: 'rangeslider',
    disabledClass: 'rangeslider--disabled',
    horizontalClass: 'rangeslider--horizontal',
    verticalClass: 'rangeslider--vertical',
    fillClass: 'rangeslider__fill',
    handleClass: 'rangeslider__handle',

    //TODO its not calling any of these callbacks. idk.
    // Callback function
    onInit: function() {
      console.log('ranger inited');
    },

    // Callback function
    onSlide: function(position, value) {
      console.log('ranger sliding"');
    },

    // Callback function
    onSlideEnd: function(position, value) {
      console.log('slide ended. doing stuff');
      $("#epsilonvalue").text(value);
      // remove all points...
      //TODO
      // getData for value as eps
      getData(map, value);
      //--> and gD does do that, namely populate points from returned q
    }
  });

});

var bounds; // google LatLngBounds
var namePositions = {}; //holds googly lat/long
var markers = []; //hold googly markers for clustering
var map; //to become a googlemap


//called from googler igniter
function initMap() {
  namePositions = {};
  markers = [];
  bounds = new google.maps.LatLngBounds();
  map = new google.maps.Map(document.getElementById('map'), {
    zoom: 3,
    center: {lat: 38.6270, lng: -90.1994},
    mapTypeId: 'terrain'
  });

    // TODO slidey mcsliderton
    getData(map,0.001)


  // addTrackPointsToMap(map);
}

function getUniqueNames(trackPoints) {
  var flags = [], output = [], l = trackPoints.length, i;
  for( i=0; i<l; i++) {
    if( flags[trackPoints[i].name]) continue;
    flags[trackPoints[i].name] = true;
    output.push(trackPoints[i].name);
  }
  return output;
}

function initNamedPositions(uniqueNames) {
  for (n in uniqueNames) {
    namePositions[uniqueNames[n]] = [];
  }
}

function getData(map,epsilon){
   return $.ajax({
     url: '/api/data/v1',
        type: 'GET',
        dataType: 'json',
        data: 'epsilon=' + parseFloat(epsilon).toString(),
        // data : JSON.stringify({ "Epsilon": epsilon}),
        success: function (data) {
          alert("got data" + JSON.stringify(data));
           addTrackPointsToMap(map,data);
       }
    });
}

function addTrackPointsToMap(map,pointsData) {

    // var pointsData = JSON.parse($("#trackPointsData").text());

    markers =[]
    if (Array.isArray(pointsData)) {
    var uniqueNames = getUniqueNames(pointsData);
    initNamedPositions(uniqueNames);

    for (var i = 0; i < pointsData.length; i++) {
      addPointMarker(map, i, pointsData[i]);
    }

    for (n in uniqueNames) {
      drawFlightPath(map, namePositions[uniqueNames[n]], uniqueNames[n]);
    }
  }else{
    alert("hey that should be array")
  }

  var markerCluster = new MarkerClusterer(map, markers, {imagePath: '/ass/images/m'});
  map.fitBounds(bounds);
}

function addPointMarker (map, index, trackPoint) {
  // console.log("adding point marker for map");
  var infoWindow = new google.maps.InfoWindow(), marker;
  var position = new google.maps.LatLng(trackPoint.lat, trackPoint.long);
  bounds.extend(position);
  namePositions[trackPoint.name].push(position);
  var markerObj = {
    position: position,
    map: map,
    title: trackPoint.name
  };
  console.log(trackPoint.name);

  switch(trackPoint.name) {
  case "jl":
    markerObj = $.extend({}, markerObj, {icon: "/ass/images/emoji/" + "water_buffalo" + ".png"});
    console.log("setting jl emoji")
    break;
  case "ia":
    markerObj = $.extend({}, markerObj, {icon: "/ass/images/emoji/" + "wolf" + ".png"});
    break;
  default:
    markerObj = $.extend({}, markerObj, {icon: "/ass/images/emoji/" + "smile" + ".png"});
  }

  //is first in namedPosition
  var isFirstIndex = namePositions[trackPoint.name].indexOf(position) == 0 ? true : false;
  if (isFirstIndex) markerObj = $.extend({}, markerObj, {animation: google.maps.Animation.BOUNCE});
  var marker = new google.maps.Marker(markerObj);

  //https://wrightshq.com/playground/placing-multiple-markers-on-a-google-map-using-api-3/
  google.maps.event.addListener(marker, 'click', (function(marker, index) {
    return function() {
      infoContent = "<h3>On this day</h3>, " + trackPoint.time + ", the cat " + trackPoint.name + " was running " + trackPoint.speed + " meters per second at an elevation of" + trackPoint.elevation + " meters";
      infoWindow.setContent(infoContent);
      infoWindow.open(map, marker);
      map.setCenter(marker.getPosition());
    }
  })(marker, index));
  // only push if not first -- because
  if (!isFirstIndex) markers.push(marker); // push to array for clustering

}

function drawFlightPath(map, positions, name) {
  var c = getRandomColor();
  var flightPath = new google.maps.Polyline({
    path: positions,
    strokeColor: c,
    strokeOpacity: 0.5,
    strokeWeight: 3
  });
  flightPath.setMap(map);
}

// http://stackoverflow.com/questions/1484506/random-color-generator-in-javascript
function getRandomColor() {
  var letters = '0123456789ABCDEF';
  var color = '#';
  for (var i = 0; i < 6; i++ ) {
    color += letters[Math.floor(Math.random() * 16)];
  }
  return color;
}
