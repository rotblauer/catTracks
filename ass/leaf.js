var eps = 0.001
var u = "/v1?" + encodeURIComponent("epsilon=" + eps);
d3.json(u, function(error, incidents) {

    console.log("incidents count: ", incidents.length);
    console.log("looking like:", incidents.splice(0, 2));

    function reformat(array) {
        var data = [];
        array.map(function(d, i) {

            if (+d.long === 0 || +d.lat === 0) return;

            data.push({
                id: i,
                type: "Feature",
                geometry: {
                    coordinates: [+d.long, +d.lat],
                    type: "Point"
                },
                properties: {
                    speed: +d.speed,
                    heading: +d.heading,
                    elevation: +d.elevation,
                    name: d.name,
                    _id: d.id,
                    accuracy: +d.accuracy,
                    tilt: +d.tilt,
                    heartrate: +d.heartrate,
                    time: d.time,
                    notes: d.notes
                }
            });
        });
        return data;
    }
    var geoData = {
        type: "FeatureCollection",
        features: reformat(incidents)
    };



    var qtree = d3.geom.quadtree(geoData.features.map(function(data, i) {
        return {
            x: data.geometry.coordinates[0],
            y: data.geometry.coordinates[1],
            all: data
        };
    }));


    // Find the nodes within the specified rectangle.
    function search(quadtree, x0, y0, x3, y3) {
        var pts = [];
        var subPixel = false;
        var subPts = [];
        var scale = getZoomScale();
        console.log(" scale: " + scale);
        var counter = 0;
        quadtree.visit(function(node, x1, y1, x2, y2) {
            var p = node.point;
            var pwidth = node.width * scale;
            var pheight = node.height * scale;

            // -- if this is too small rectangle only count the branch and set opacity
            if ((pwidth * pheight) <= 1) {
                // start collecting sub Pixel points
                subPixel = true;
            }
            // -- jumped to super node large than 1 pixel
            else {
                // end collecting sub Pixel points
                if (subPixel && subPts && subPts.length > 0) {

                    subPts[0].group = subPts.length;
                    pts.push(subPts[0]); // add only one todo calculate intensity
                    counter += subPts.length - 1;
                    subPts = [];
                }
                subPixel = false;
            }

            if ((p) && (p.x >= x0) && (p.x < x3) && (p.y >= y0) && (p.y < y3)) {

                if (subPixel) {
                    subPts.push(p.all);
                } else {
                    if (p.all.group) {
                        delete(p.all.group);
                    }
                    pts.push(p.all);
                }

            }
            // if quad rect is outside of the search rect do nto search in sub nodes (returns true)
            return x1 >= x3 || y1 >= y3 || x2 < x0 || y2 < y0;
        });
        console.log(" Number of removed  points: " + counter);
        return pts;

    }


    function updateNodes(quadtree) {
        var nodes = [];
        quadtree.depth = 0; // root

        quadtree.visit(function(node, x1, y1, x2, y2) {
            var nodeRect = {
                left: MercatorXofLongitude(x1),
                right: MercatorXofLongitude(x2),
                bottom: MercatorYofLatitude(y1),
                top: MercatorYofLatitude(y2),
            };
            node.width = (nodeRect.right - nodeRect.left);
            node.height = (nodeRect.top - nodeRect.bottom);

            if (node.depth == 0) {
                console.log(" width: " + node.width + "height: " + node.height);
            }
            nodes.push(node);
            for (var i = 0; i < 4; i++) {
                if (node.nodes[i]) node.nodes[i].depth = node.depth + 1;
            }
        });
        return nodes;
    }

    //-------------------------------------------------------------------------------------
    MercatorXofLongitude = function(lon) {
        return lon * 20037508.34 / 180;
    }

    MercatorYofLatitude = function(lat) {
        return (Math.log(Math.tan((90 + lat) * Math.PI / 360)) / (Math.PI / 180)) * 20037508.34 / 180;
    }
    var cscale = d3.scale.linear().domain([1, 3]).range(["#ff0000", "#ff6a00", "#ffd800", "#b6ff00", "#00ffff", "#0094ff"]); //"#00FF00","#FFA500"

    var leafletMap = L.map('map').setView([13.4, 52.5], 2);

    var mapstack_original;
    var mapstack_water1 = "http://mapstack.stamen.com/edit.html#watercolor[sat=40,comp=lighter,alpha=70]/11/37.7550/-122.3513";
    var ms_water1_url = "http://{s}.tile.mapstack.stamen.com/(watercolor,$ff[lighter],$fff[sat@40],$fff[alpha@70])/{z}/{x}/{y}.png";
    var tile_ex = "http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"; //attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
    var os_tile_bw = "http://korona.geog.uni-heidelberg.de/tiles/roadsg/x={x}&y={y}&z={z}";
    // L.tileLayer("http://{s}.sm.mapstack.stamen.com/(toner-lite,$fff[difference],$fff[@23],$fff[hsl-saturation@20])/{z}/{x}/{y}.png").addTo(leafletMap);
    L.tileLayer(os_tile_bw).addTo(leafletMap);

    var svg = d3.select(leafletMap.getPanes().overlayPane).append("svg");
    var g = svg.append("g").attr("class", "leaflet-zoom-hide");


    // Use Leaflet to implement a D3 geometric transformation.
    function projectPoint(x, y) {
        var point = leafletMap.latLngToLayerPoint(new L.LatLng(y, x));
        this.stream.point(point.x, point.y);
    }

    var transform = d3.geo.transform({
        point: projectPoint
    });
    var path = d3.geo.path().projection(transform);


    updateNodes(qtree);

    leafletMap.on('moveend', mapmove);

    function fitMapToAllPoints() {
        var arrayOfLatLngs = [];
        geoData.features.map(function(d, i) {
          //the world is upsidedown
          arrayOfLatLngs.push([d.geometry.coordinates[1], d.geometry.coordinates[0]]);
        });
        console.log("aoll", arrayOfLatLngs);
        var bs = new L.LatLngBounds(arrayOfLatLngs);
        console.log("bs", bs);
      leafletMap.fitBounds([bs.getNorthWest(), bs.getSouthEast()]);
      leafletMap.panTo(bs.getCenter());
      mapmove();
    }
    fitMapToAllPoints();

    function getZoomScale() {
        var mapWidth = leafletMap.getSize().x;
        var bounds = leafletMap.getBounds();
        var planarWidth = MercatorXofLongitude(bounds.getEast()) - MercatorXofLongitude(bounds.getWest());
        var zoomScale = mapWidth / planarWidth;
        return zoomScale;

    }

    function redrawSubset(subset) {
        if (subset.length === 0) return;
        path.pointRadius(3); // * scale);

        var bounds = path.bounds({
            type: "FeatureCollection",
            features: subset
        });
        var topLeft = bounds[0];
        var bottomRight = bounds[1];


        svg.attr("width", bottomRight[0] - topLeft[0])
            .attr("height", bottomRight[1] - topLeft[1])
            .style("left", topLeft[0] + "px")
            .style("top", topLeft[1] + "px");


        g.attr("transform", "translate(" + -topLeft[0] + "," + -topLeft[1] + ")");

        var start = new Date();


        var points = g.selectAll("path")
            .data(subset, function(d) {
                return d.id;
            });
        points.enter().append("path");
        points.exit().remove();
        points.attr("d", path);

        points.style("fill", function(d) {
            if (d.properties.name === "Big Papa" || d.properties.name === "ia") return "red";
            if (d.properties.name === "RyePhone" || d.properties.name === "jl") return "blue";
            if (d.properties.name === "Big Mamma") return "green";
            return "yellow";
        });
        points.style("fill-opacity", function(d) {
            if (d.group) {
                return (d.group * 0.1) + 0.2;
            }
        });


        console.log("updated at  " + new Date().setTime(new Date().getTime() - start.getTime()) + " ms ");

    }

    var counter = 0;

    function mapmove(e) {
        var mapBounds = leafletMap.getBounds();
        console.log('mapbounds', mapBounds);

        var subset = search(qtree, mapBounds.getWest(), mapBounds.getSouth(), mapBounds.getEast(), mapBounds.getNorth());
        console.log("subset.length: " + subset.length);

        redrawSubset(subset);
        if (counter === 0) {
            leafletMap.fit
        }
        counter++;
    }

});
