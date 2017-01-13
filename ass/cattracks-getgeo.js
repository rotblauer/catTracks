
     geolocator.config({
         language: "en",
         google: {
             version: "3",
             key: "AIzaSyCEaBxQ_ghjyTOfe61WiWjPk8FCQPbMJuo"
         }
     });


var getGeo = function () {
         $("#pleasehold").text("kindly still getting your location madam");
         var options = {
             enableHighAccuracy: true,
             timeout: 5000,
             maximumWait: 10000,     // max wait time for desired accuracy
             maximumAge: 0,          // disable cache
             desiredAccuracy: 30,    // meters
             fallbackToIP: true,     // fallback to IP if Geolocation fails or rejected
             /* addressLookup: true,    // requires Google API key if true*/
             timezone: true         // requires Google API key if true
             /* map: "map-canvas",      // interactive map element id (or options object)*/
             /* staticMap: true         // map image URL (boolean or options object)*/
         };
         geolocator.locate(options, function (err, location) {
             if (err) return console.log(err);
             // console.log(location);
           console.log("got location...");
           gotLocation(location);
           $("#pleasehold").text("");


             /* Object*/
             /* coords: Object*/
             /* accuracy: 65*/
             /* altitude: null*/
             /* altitudeAccuracy: null*/
             /* heading: null*/
             /* latitude: 52.484596951840125*/
             /* longitude: 13.445309906232552*/
             /* speed: null*/
             /* Object Prototype*/
             /* timestamp: 1484300665425*/
             /* timezone: Object*/
             /* abbr: "CEST"*/
             /* dstOffset: 0*/
             /* id: "Europe/Berlin"*/
             /* name: "Central European Standard Time"*/
             /* rawOffset: 3600*/
             /* Object Prototype*/
             /* Object Prototype*/

         });
     };
// on getting location, fill in the form
function gotLocation (location) {
  console.log("Got location: ", location);
  currentLocation = location;
  $('#myLat').val(location.coords.latitude);
  $("#myLon").val(location.coords.longitude);

}



window.onload = getGeo;
