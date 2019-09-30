function loadSelected() {
	var rq = {
		areas: selectionCtl.selected
	}

	if (selectionCtl.pointName) {
		rq.point = {
			id: Math.floor(Math.random() * 1000000),
			name: selectionCtl.pointName,
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
	selectionCtl.clearSelection()
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
	areasCtl.dropArea(area)
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
	constructor(states, cb) {
		this.states = states
		this.cb = cb
		this.reset()
	}

	toggle() {
		this.i = (this.i + 1) % this.states.length
		this.text = this.states[this.i]

		if (this.cb != null) {
			this.cb()
		}
	}

	reset() {
		this.i = 0
		this.text = this.states[0]
	}
}

var showTypesToggle = new toggle(["more", "less"])

showTypesToggle.show = function(area) {
	return showTypesToggle.i == 1 || area.type == "O02" || area.type == "O04"
}

var hidePointsToggle = new toggle(["hide", "show"],
	function() {
		if (hidePointsToggle.i == 0) {
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

var areasCtl = new Vue({
	el: '#areas',
	data: {
		loaded: {},
		nr: 0,
	},
	methods: {
		addArea: (area) => {
			areasCtl.$set(areasCtl.loaded, area.id, area)
			areasCtl.nr += 1
		},

		updateArea: (area, layer) => {
			area.state = "ready"
			area.layer = layer
			areasCtl.$set(areasCtl.loaded, area.id, area)
		},

		dropArea: (area) => {
			areasCtl.$delete(areasCtl.loaded, area.id)
			areasCtl.nr -= 1
		},

	},
})

var selectionCtl = new Vue({
	el: '#selection',
	data: {
		available: [],
		selected: [],
		pointName: "",

		show: showTypesToggle,
	},
	methods: {
		move: (latlng) => {
			selectionCtl.clearSelection()
		},

		clearSelection: () => {
			selectionCtl.pointName = ""
			selectionCtl.available = []
			selectionCtl.selected = []
			selectionCtl.show.reset()
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

			selectionCtl.available = areas
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
	selectionCtl.move(e.latlng)

	reqwest({
		url: 'https://global.mapit.mysociety.org/point/4326/'+e.latlng.lng+','+e.latlng.lat,
		method: 'GET',
		crossOrigin: true,
		success: (data) => {
			selectionCtl.setAvailable(data)
		},
		error: (e) => {
			mymarker.remove()
			selectionCtl.clearSelection()
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
	areasCtl.updateArea(this.area, L.stamp(area))

	console.log("Loaded " + this.area.name)
};

myareas.addArea = function(area) {
	if (!areasCtl.loaded[area.id]) {
		console.log("Requesting ", area.name);
		areasCtl.addArea(area)

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
