function loadSelected() {
	reqwest({
		url: apiserver + '/visited',
		method: 'POST',
		contentType: 'application/json',
		data: JSON.stringify({
			lat: areaCtl.point.lat,
			lng: areaCtl.point.lng,
			areas: areaCtl.selectedAreas
		}),
		crossOrigin: true,
		success: (x) => {
			console.log("Added to backend");
		},
	})

	areaCtl.selectedAreas.forEach((item, i) => { myareas.load(item) })

	clickPoint.remove()
	areaCtl.clearSelection()
}

function removeLoaded(ev, area) {
	reqwest({
			url: apiserver + '/visited?id=' + area.id,
			method: 'DELETE',
			crossOrigin: true,
			success: (x) => {
				console.log("Removed from backend")
			},
	})

	myareas.loaded.removeLayer(area.layer)
	areaCtl.dropLoaded(area)
}

var areaCtl = new Vue({
	el: '#control',
	data: {
		point: null,
		availableAreas: null,
		selectedAreas: [],
		loadedAreas: {},
		nrLoaded: 0,
	},
	methods: {
		clearSelection: () => {
			areaCtl.point = null
			areaCtl.availableAreas = null
			areaCtl.selectedAreas = []
		},

		move: (latlng) => {
			areaCtl.clearSelection()
			areaCtl.point = latlng
		},

		addLoaded: (area, layer) => {
			areaCtl.$set(areaCtl.loadedAreas, area.id, {
					id: area.id,
					name: area.name,
					layer: layer,
				})
			areaCtl.nrLoaded += 1
		},

		dropLoaded: (area) => {
			areaCtl.$delete(areaCtl.loadedAreas, area.id)
			areaCtl.nrLoaded -= 1
		},
	}
})

var mymap = L.map('map').setView([53.505, 25.09], 5);

var osm = new L.TileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
	attribution: 'Map © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	maxZoom: 18
    });
mymap.addLayer(osm);

var clickPoint = clickPoint || {}

clickPoint.marker = null
clickPoint.move = function(e) {
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
	clickPoint.move(e)
	areaCtl.move(e.latlng)

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
		"color": areaColor,
	};
	var area = new L.GeoJSON(data, { style: style });
	area.on('dblclick', function(e){
	    var z = mymap.getZoom() + (e.originalEvent.shiftKey ? -1 : 1);
	    mymap.setZoomAround(e.containerPoint, z);
	});

	myareas.loaded.addLayer(area);
	areaCtl.addLoaded(this.myareas, L.stamp(area))

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
		url: apiserver + '/visited',
		method: 'GET',
		type: 'json',
		crossOrigin: true,
		success: (data) => {
			console.log('Visited areas: ');
			console.log(data);
		},
})
