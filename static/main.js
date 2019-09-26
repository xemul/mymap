function loadSelected() {
	reqwest({
		url: '/visited',
		method: 'POST',
		contentType: 'application/json',
		data: JSON.stringify(areaCtl.selectedAreas),
		crossOrigin: true,
		success: (x) => {
			console.log("Added to backend");
		},
	})

	areaCtl.selectedAreas.forEach((item, i) => { myareas.load(item) })
	areaCtl.selectedAreas = [];

	clickPoint.marker.remove()
	areaCtl.point = null
	areaCtl.availableAreas = null
}

function removeLoaded(ev, area) {
	reqwest({
			url: '/visited?id=' + area.id,
			method: 'DELETE',
			crossOrigin: true,
			success: (x) => {
				console.log("Removed from backend")
			},
	})

	myareas.loaded.removeLayer(area.layer)
	areaCtl.$delete(areaCtl.loadedAreas, area.id)
	areaCtl.nrLoaded -= 1
}

var areaCtl = new Vue({
	el: '#point',
	data: {
		point: null,
		availableAreas: null,
		selectedAreas: [],
		loadedAreas: {},
		nrLoaded: 0,
	},
})

var mymap = L.map('map').setView([53.505, 25.09], 5);

var osm = new L.TileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
	attribution: 'Map © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	maxZoom: 18
    });
mymap.addLayer(osm);

var clickPoint = clickPoint || {}

clickPoint.marker = null
clickPoint.add = function(e) {
	if (clickPoint.marker != null) {
		clickPoint.marker.remove()
	}
	clickPoint.marker = L.marker(e.latlng).addTo(mymap);
}

clickPoint.remove = function() {
	if (clickPoint.marker != null) {
		clickPoint.marker.remove()
		clickPoint.marker = null
	}
}

mymap.on('click', (e) => {
	clickPoint.add(e)
	areaCtl.point = '' + e.latlng.lat.toFixed(4) + ', ' + e.latlng.lng.toFixed(4);

	reqwest({
		url: 'https://global.mapit.mysociety.org/point/4326/'+e.latlng.lng+','+e.latlng.lat,
		method: 'GET',
		crossOrigin: true,
		success: (data) => {
			areaCtl.availableAreas = data;
		},
	})
})

var myareas = myareas || {};
myareas.loaded = L.featureGroup().addTo(mymap);

myareas.area_loaded = function(data) {
	var style = {
		"weight": 0.01,
		"color": "#4041ff",
	};
	var area = new L.GeoJSON(data, { style: style });
	area.on('dblclick', function(e){
	    var z = mymap.getZoom() + (e.originalEvent.shiftKey ? -1 : 1);
	    mymap.setZoomAround(e.containerPoint, z);
	});

	myareas.loaded.addLayer(area);
	areaCtl.$set(areaCtl.loadedAreas, this.myareas.id, {
			id: this.myareas.id,
			name: this.myareas.name,
			layer: L.stamp(area),
		})
	areaCtl.nrLoaded += 1

	console.log("Loaded " + this.myareas.name)
};

myareas.load = function(item) {
	if (!areaCtl.loadedAreas[item.id]) {
		console.log("Requesting ", item.name);
		reqwest({
			url: 'https://global.mapit.mysociety.org/area/' + item.id + '.geojson?simplify_tolerance=0.0001',
			type: 'json',
			myareas: { id: item.id, name: item.name },
			success: myareas.area_loaded,
			crossOrigin: true
		});
	}
}

reqwest({
		url: '/visited',
		method: 'GET',
		type: 'json',
		crossOrigin: true,
		success: (data) => {
			console.log('Visited areas: ');
			console.log(data);
		},
})
