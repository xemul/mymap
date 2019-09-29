function loadSelected() {
	var rq = {
		areas: areaCtl.selectedAreas
	}

	if (areaCtl.pointName) {
		rq.point = {
			id: Math.floor(Math.random() * 1000000),
			name: areaCtl.pointName,
			lat: markerCtl.latlng.lat,
			lng: markerCtl.latlng.lng,
			countries: markerCtl.countries,
		}
	}

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
		mypoints.addPoint(rq.point)
	}

	mymarker.remove()
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

	mypoints.loaded.removeLayer(pnt.marker)
	pointsCtl.dropPoint(pnt)
}

class toggle {
	constructor(tOn, tOff, cb) {
		this.on = tOn
		this.off = tOff
		this.reset()
		this.cb = cb
	}

	toggle() {
		if (this.state) {
			this.state = false
			this.text = this.off
		} else {
			this.state = true
			this.text = this.on
		}

		if (this.cb != null) {
			this.cb()
		}
	}

	reset() {
		this.state = false
		this.text = this.off
	}
}

var showTypesToggle = new toggle("less", "more")

showTypesToggle.show = function(area) {
	return showTypesToggle.state || area.type == "O02" || area.type == "O04"
}

var hidePointsToggle = new toggle("show", "hide",
	function() {
		if (!hidePointsToggle.state) {
			mymap.addLayer(mypoints.loaded)
		} else {
			mymap.removeLayer(mypoints.loaded)
		}
	}
)

var loginCtl = new Vue({
	el: '#login',
	data: {
		status: "Not logged in",
	},
})

var markerCtl = new Vue({
	el: '#marker',
	data: {
		latlng: null,
		countries: [],
	},
})

var pointsCtl = new Vue({
	el: '#points',
	data: {
		loaded: {},
		nr: 0,
		hide: hidePointsToggle,
	},
	methods: {
		addPoint: (pt) => {
			let bkey = pt.countries.join(',')
			var bucket = pointsCtl.loaded[bkey]

			if (!bucket) {
				pointsCtl.$set(pointsCtl.loaded, bkey, {})
				bucket = pointsCtl.loaded[bkey]
			}

			Vue.set(bucket, pt.id, pt)
			pointsCtl.nr += 1
		},

		dropPoint: (pt) => {
			let bkey = pt.countries.join(',')
			var bucket = pointsCtl.loaded[bkey]

			Vue.delete(bucket, pt.id)

			if (Object.keys(bucket).length == 0) {
				pointsCtl.$delete(pointsCtl.loaded, bkey)
			}
			pointsCtl.nr -= 1
		},

	},
})

var areaCtl = new Vue({
	el: '#control',
	data: {
		pointName: "",
		availableAreas: [],
		showTypes: showTypesToggle,
		selectedAreas: [],
		loadedAreas: {},
		nrAreas: 0,
	},
	methods: {
		move: (latlng) => {
			areaCtl.clearSelection()
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

		clearSelection: () => {
			areaCtl.pointName = ""
			areaCtl.availableAreas = []
			areaCtl.selectedAreas = []
			areaCtl.showTypes.reset()
		},

		setAvailable: (data) => {
			let areas = []
			let countries = []
			let last = null

			Object.keys(data).forEach((key, idx) => {
				var area = data[key]
				areas.push({
					id:	area.id,
					name:	area.name,
					state:	"loading",
					type:	area.type,
				})

				last = area
			})
			if (last && last.countries) {
				last.countries.forEach((item, i) => { countries.push(item.code) })
			}
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
			markerCtl.countries = countries
		},
	}
})

var mymap = L.map('map').setView([53.505, 25.09], 5);

var osm = new L.TileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
	attribution: 'Map © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	maxZoom: 18
    });
mymap.addLayer(osm);

var mymarker = mymarker || {}
mymarker.lm = null

mymarker.move = function(latlng) {
	if (mymarker.lm != null) {
		mymarker.lm.remove()
	}
	mymarker.lm = L.marker(latlng, {icon: pointIcon}).addTo(mymap);
	markerCtl.latlng = latlng
	markerCtl.countries = []
}

mymarker.remove = function() {
	if (mymarker.lm != null) {
		mymarker.lm.remove()
		mymarker.lm = null
	}

	markerCtl.latlng = null
	markerCtl.countries = []
}

mymap.on('click', (e) => {
	mymarker.move(e.latlng)
	areaCtl.move(e.latlng)

	reqwest({
		url: 'https://global.mapit.mysociety.org/point/4326/'+e.latlng.lng+','+e.latlng.lat,
		method: 'GET',
		crossOrigin: true,
		success: (data) => {
			areaCtl.setAvailable(data)
		},
		error: (e) => {
			mymarker.remove()
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

var mypoints = mypoints || {}
mypoints.loaded = L.layerGroup().addTo(mymap);

mypoints.addPoint = function(pt) {
	pt.marker = L.marker(pt, {icon: placeIcon}).addTo(mypoints.loaded)
	pt.marker.bindTooltip(pt.name, {direction: "auto", opacity: placeTolltipOpacity})
	pointsCtl.addPoint(pt)
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
					mypoints.addPoint(item)
				})
			}
		},
})
