function loadSelected() {
	var rq = {
		areas: areaCtl.selectedAreas
	}

	if (areaCtl.pointName) {
		rq.point = {
			id: Math.floor(Math.random() * 1000000),
			name: areaCtl.pointName,
			lat: areaCtl.point.lat,
			lng: areaCtl.point.lng,
		}
	}

	console.log("save:", areaCtl.pointName)
	reqwest({
		url: apiserver + '/visited',
		method: 'POST',
		contentType: 'application/json',
		data: JSON.stringify(rq),
		crossOrigin: true,
		success: (x) => {
			console.log("Added to backend");
		},
	})

	rq.areas.forEach((item, i) => { myareas.addArea(item) })
	if (rq.point) {
		myareas.addPoint(rq.point)
	}

	clickPoint.remove()
	areaCtl.clearSelection()
}

function removeArea(ev, area) {
	reqwest({
			url: apiserver + '/visited?type=area&id=' + area.id,
			method: 'DELETE',
			crossOrigin: true,
			success: (x) => {
				console.log("Removed area from backend")
			},
	})

	myareas.loaded.removeLayer(area.layer)
	areaCtl.dropArea(area)
}

function removePoint(ev, pnt) {
	reqwest({
			url: apiserver + '/visited?type=point&id=' + pnt.id,
			method: 'DELETE',
			crossOrigin: true,
			success: (x) => {
				console.log("Removed point from backend")
			},
	})

	mymap.removeLayer(pnt.marker)
	areaCtl.dropPoint(pnt)
}

var areaCtl = new Vue({
	el: '#control',
	data: {
		point: null,
		pointName: "",
		availableAreas: [],
		selectedAreas: [],
		loadedAreas: {},
		nrAreas: 0,
		loadedPoints: {},
		nrPoints: 0,
	},
	methods: {
		move: (latlng) => {
			areaCtl.clearSelection()
			areaCtl.point = latlng
		},

		addArea: (area) => {
			areaCtl.$set(areaCtl.loadedAreas, area.id, area)
			areaCtl.nrAreas += 1
		},

		updateArea: (area, layer) => {
			area.state = "ready"
			area.layer = layer
			areaCtl.$set(areaCtl.loadedAreas, area.id, area)
		},

		dropArea: (area) => {
			areaCtl.$delete(areaCtl.loadedAreas, area.id)
			areaCtl.nrAreas -= 1
		},

		addPoint: (pt) => {
			areaCtl.$set(areaCtl.loadedPoints, pt.name, pt)
			areaCtl.nrPoints += 1
		},

		dropPoint: (pt) => {
			areaCtl.$delete(areaCtl.loadedPoints, pt.name)
			areaCtl.nrPoints -= 1
		},

		clearSelection: () => {
			areaCtl.point = null
			areaCtl.pointName = ""
			areaCtl.availableAreas = []
			areaCtl.selectedAreas = []
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
	areaCtl.updateArea(this.area, L.stamp(area))

	console.log("Loaded " + this.area.name)
};

myareas.addArea = function(area) {
	if (!areaCtl.loadedAreas[area.id]) {
		console.log("Requesting ", area.name);
		areaCtl.addArea(area)

		reqwest({
			url: 'https://global.mapit.mysociety.org/area/' + area.id + '.geojson?simplify_tolerance=0.0001',
			type: 'json',
			area: area,
			success: myareas.area_loaded,
			crossOrigin: true
		});
	}
}

myareas.addPoint = function(pt) {
	pt.marker = L.marker(pt, {icon: placeIcon}).addTo(mymap)
	areaCtl.addPoint(pt)
}

reqwest({
		url: apiserver + '/visited',
		method: 'GET',
		type: 'json',
		crossOrigin: true,
		success: (data) => {
			if (data.areas) {
				data.areas.forEach((item, i) => {
					item.state = "loading"
					myareas.addArea(item)
				})
			}
			if (data.points) {
				data.points.forEach((item, i) => {
					myareas.addPoint(item)
				})
			}
		},
})
