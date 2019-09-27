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
	L.marker(areaCtl.point, { icon: placeIcon }).addTo(mymap)

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
		availableAreas: [],
		selectedAreas: [],
		loadedAreas: {},
		nrLoaded: 0,
	},
	methods: {
		clearSelection: () => {
			areaCtl.point = null
			areaCtl.availableAreas = []
			areaCtl.selectedAreas = []
		},

		move: (latlng) => {
			areaCtl.clearSelection()
			areaCtl.point = latlng
		},

		addLoaded: (area) => {
			areaCtl.$set(areaCtl.loadedAreas, area.id, area)
			areaCtl.nrLoaded += 1
		},

		updateLoaded: (area, layer) => {
			area.state = "ready"
			area.layer = layer
			areaCtl.$set(areaCtl.loadedAreas, area.id, area)
		},

		dropLoaded: (area) => {
			areaCtl.$delete(areaCtl.loadedAreas, area.id)
			areaCtl.nrLoaded -= 1
		},

		setAvailable: (data) => {
			let areas = []
			Object.keys(data).forEach((key, idx) => {
				var area = data[key]
				areas.push({
					id:	area.id,
					name:	area.name,
					state:	"loading",
					type:	area.type,
				})
			})
			areas.sort((a,b)=>{
				if (a.type > b.type) {
					return 1
				} else if (a.type < b.type) {
					return -1
				} else {
					return 0
				}
			})
			areaCtl.availableAreas = areas
		},
	}
})

var mymap = L.map('map').setView([53.505, 25.09], 5);

var osm = new L.TileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
	attribution: 'Map © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	maxZoom: 18
    });
mymap.addLayer(osm);

const pointIcon = L.icon({
	iconUrl:	'img/point.svg',
	iconSize:	[30, 48],
	iconAnchor:	[15, 47],
})

const placeIcon = L.icon({
	iconUrl:	'img/place.svg',
	iconSize:	[14, 21],
	iconAnchor:	[ 7, 20],
})

var clickPoint = clickPoint || {}
clickPoint.marker = null

clickPoint.move = function(latlng) {
	if (clickPoint.marker != null) {
		clickPoint.marker.remove()
	}
	clickPoint.marker = L.marker(latlng, {icon: pointIcon}).addTo(mymap);
}

clickPoint.remove = function() {
	if (clickPoint.marker != null) {
		clickPoint.marker.remove()
		clickPoint.marker = null
	}
}

mymap.on('click', (e) => {
	clickPoint.move(e.latlng)
	areaCtl.move(e.latlng)

	reqwest({
		url: 'https://global.mapit.mysociety.org/point/4326/'+e.latlng.lng+','+e.latlng.lat,
		method: 'GET',
		crossOrigin: true,
		success: (data) => {
			areaCtl.setAvailable(data)
		},
		error: (e) => {
			clickPoint.remove()
			areaCtl.clearSelection()
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

	myareas.loaded.addLayer(area)
	areaCtl.updateLoaded(this.area, L.stamp(area))

	console.log("Loaded " + this.area.name)
};

myareas.load = function(area) {
	if (!areaCtl.loadedAreas[area.id]) {
		console.log("Requesting ", area.name);
		areaCtl.addLoaded(area)

		reqwest({
			url: 'https://global.mapit.mysociety.org/area/' + area.id + '.geojson?simplify_tolerance=0.0001',
			type: 'json',
			area: area,
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
			data.areas.forEach((item, i) => {
				item.state = "loading"
				myareas.load(item)
			})
			data.points.forEach((item, i) => {
				L.marker([item.lat, item.lng], {icon: placeIcon}).addTo(mymap);
			})
		},
})
