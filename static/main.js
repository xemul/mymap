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
			countries: markerCtl.inside.countries,
			area: markerCtl.inside.area,
		}
	}

	reqwest({
		url: apiserver + '/geos',
		method: 'POST',
		contentType: 'application/json',
		data: JSON.stringify(rq),
		crossOrigin: true,
		success: (x) => {
			console.log("Added to backend");
		},
	})

	rq.areas.forEach((item, i) => { areasLayer.addArea(item) })
	if (rq.point) {
		pointsLayer.addPoint(rq.point)
	}

	markerLayer.remove()
	selectionCtl.clearSelection()
}

function removeArea(ev, area) {
	reqwest({
			url: apiserver + '/geos?type=area&id=' + area.id,
			method: 'DELETE',
			crossOrigin: true,
			success: (x) => {
				console.log("Removed area from backend")
			},
	})

	areasLayer.loaded.removeLayer(area.layer)
	areasCtl.dropArea(area)
}

function removePoint(ev, pnt) {
	reqwest({
			url: apiserver + '/geos?type=point&id=' + pnt.id,
			method: 'DELETE',
			crossOrigin: true,
			success: (x) => {
				console.log("Removed point from backend")
			},
	})

	pointsLayer.loaded.removeLayer(pnt.marker)
	pointsCtl.dropPoint(pnt)
}

function highlightPoint(ev, pnt) {
	mymap.setView(pnt, highlightZoom)
	pnt.marker.setIcon(placeDIcon)
	setTimeout(() => { pnt.marker.setIcon(placeIcon) }, highlightTimeout)
}

function clearMarker(ev) {
	markerLayer.remove()
	selectionCtl.clearSelection()
}

class toggle {
	constructor(states, cb) {
		this.states = states
		this.cb = cb
		this.reset()
	}

	toggle() {
		this.i = (this.i + 1) % this.states

		if (this.cb != null) {
			this.cb()
		}
	}

	reset() {
		this.i = 0
	}
}

var showTypesToggle = new toggle(2)

showTypesToggle.show = function(area) {
	return showTypesToggle.i == 1 || area.type == "O02" || area.type == "O04"
}

var hidePointsToggle = new toggle(2,
	function() {
		if (hidePointsToggle.i == 0) {
			mymap.addLayer(pointsLayer.loaded)
		} else {
			mymap.removeLayer(pointsLayer.loaded)
		}
	}
)

var hideAreasToggle = new toggle(2)

//
// Sidebar stuff
//

var menuCtl = new Vue({
	el: '#menu',
	data: {
		sess: null,
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
		},

		setAvailable: (data) => {
			let areas = []
			let smallest = null

			Object.keys(data).forEach((key, idx) => {
				var area = data[key]
				areas.push({
					id:	area.id,
					name:	area.name,
					state:	"loading",
					type:	area.type,
				})

				if (smallest == null || smallest.type > area.type) {
					smallest = area
				}
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

			let inside = { countries: [] }
			if (smallest) {
				if (smallest.countries) {
					smallest.countries.forEach((item, i) => { inside.countries.push(item.code) })
				}
				inside.area = smallest.id
			}

			selectionCtl.available = areas
			markerCtl.inside = inside
		},
	}
})

var mapCtl = new Vue({
	el: '#map',
	data: {
		width:	mapWidth,
		height:	mapHeight,
	},
	methods: {
		resize: (neww, newh, pt) => {
			mapCtl.width = neww
			mapCtl.height = newh
			Vue.nextTick(() => {
				mymap.invalidateSize()
				if (pt != null) {
					highlightPoint(null, pt)
				}
			})
		}
	},
})

var mymap = L.map('map').setView(startAt, 5);

var osm = new L.TileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
	attribution: 'Map © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
	maxZoom: 18
    });
mymap.addLayer(osm);

mymap.on('click', (e) => {
	markerLayer.move(e.latlng)
	selectionCtl.move(e.latlng)

	reqwest({
		url: 'https://global.mapit.mysociety.org/point/4326/'+e.latlng.lng+','+e.latlng.lat,
		method: 'GET',
		crossOrigin: true,
		success: (data) => {
			selectionCtl.setAvailable(data)
		},
		error: (e) => {
			clearMarker()
		},
	})
})

//
// Marker
//

var markerLayer = markerLayer || {}
markerLayer.lm = null

markerLayer.move = function(latlng) {
	if (markerLayer.lm != null) {
		markerLayer.lm.remove()
	}
	markerLayer.lm = L.marker(latlng, {icon: pointIcon}).addTo(mymap);
	markerCtl.latlng = latlng
	markerCtl.inside = null
}

markerLayer.remove = function() {
	if (markerLayer.lm != null) {
		markerLayer.lm.remove()
		markerLayer.lm = null
	}

	markerCtl.latlng = null
	markerCtl.inside = null
}

var markerCtl = new Vue({
	el: '#marker',
	data: {
		latlng: null,
		inside: [],
	},
})

//
// Areas
//

var areasLayer = areasLayer || {};
areasLayer.loaded = L.featureGroup().addTo(mymap);

areasLayer.area_loaded = function(data) {
	var style = {
		"weight": 0.01,
		"color": areaColor,
	};
	var area = new L.GeoJSON(data, { style: style });
	area.on('dblclick', function(e) {
	    var z = mymap.getZoom() + (e.originalEvent.shiftKey ? -1 : 1);
	    mymap.setZoomAround(e.containerPoint, z);
	});

	areasLayer.loaded.addLayer(area)
	areasCtl.updateArea(this.area, L.stamp(area))

	console.log("Loaded " + this.area.name)
};

areasLayer.addArea = function(area) {
	if (!areasCtl.loaded[area.id]) {
		console.log("Requesting ", area.name);
		areasCtl.addArea(area)

		reqwest({
			url: 'https://global.mapit.mysociety.org/area/' + area.id + '.geojson?simplify_tolerance=0.0001',
			type: 'json',
			area: area,
			success: areasLayer.area_loaded,
			crossOrigin: true
		});
	}
}

var areasCtl = new Vue({
	el: '#areas',
	data: {
		loaded: {},
		nr: 0,
		hide: hideAreasToggle,
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

//
// Points
//

var pointsLayer = pointsLayer || {}
pointsLayer.loaded = L.layerGroup().addTo(mymap);

pointsLayer.addPoint = function(pt) {
	pt.marker = L.marker(pt, {icon: placeIcon}).addTo(pointsLayer.loaded)
	pt.marker.bindTooltip(pt.name, {direction: "auto", opacity: placeTolltipOpacity})
	pt.marker.on('click', function(e) {
		showPropPoint(pt)
	})
	pointsCtl.addPoint(pt)
}

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

//
// Props
//

function showPropTimeline() {
	propsCtl.timeline = true
	mapCtl.resize("50%", mapHeight)
}

function clearPropTimeline() {
	propsCtl.timeline = false
	growPropMap()
}

function showPropPoint(pt) {
	propsCtl.showPoint(pt)
	mapCtl.resize("50%", mapHeight, pt)
}

function clearPropPoint(e) {
	propsCtl.clearPoint()
	growPropMap()
}

function growPropMap() {
	if (!propsCtl.timeline && propsCtl.point == null) {
		mapCtl.resize(mapWidth, mapHeight, null)
	}
}

var propsCtl = new Vue({
	el: '#props',
	data: {
		point: null,
		visited: [],
		nvDate: "",
		nvTags: "",

		timeline: false,
	},
	methods: {
		clearPoint: () => {
			propsCtl.point = null
			propsCtl.visited = []
			propsCtl.clearNv()
		},

		clearNv: () => {
			propsCtl.nvDate = ""
			propsCtl.nvTags = ""
		},

		showPoint: (pt) => {
			propsCtl.point = pt
			propsCtl.visited = []
			propsCtl.clearNv()

			reqwest({
				url: apiserver + '/visits?id=' + pt.id,
				method: 'GET',
				type: 'json',
				crossOrigin: true,
				success: (data) => {
					if (data.array) {
						data.array.forEach((v, i) => {
							v.idx = i
							propsCtl.visited.push(v)
						})
						propsCtl.sortVisited()
					}
				},
			})
		},

		commit: (nv) => {
			nv.idx = propsCtl.visited.length
			propsCtl.visited.push(nv)
			propsCtl.sortVisited()
			propsCtl.clearNv()
		},

		dropVisit: (di) => {
			let rdi = propsCtl.visited[di].idx
			propsCtl.visited.splice(di, 1)
			propsCtl.visited.forEach((v, i) => {
				if (v.idx > rdi) {
					v.idx -= 1
				}
			})
		},

		sortVisited: () => {
			propsCtl.visited.sort((a,b) => {
				return dateScore(b.date) - dateScore(a.date)
			})
		},
	}
})

function dateScore(date) {
	let ds = date.split("/")

	let y = parseInt(ds.pop()) || 0
	let m = parseInt(ds.pop()) || 0
	let d = parseInt(ds.pop()) || 0

	return ((y * 12) + m) * 32 + d
}

function addVisit() {
	let nv = {
		date: propsCtl.nvDate,
		tags: propsCtl.nvTags.split(/\s*,\s*/),
	}

	reqwest({
		url: apiserver + '/visits?id=' + propsCtl.point.id,
		method: 'POST',
		contentType: 'application/json',
		data: JSON.stringify(nv),
		crossOrigin: true,
		success: (data) => {
			propsCtl.commit(nv)
		},
		error: (err) => {
			console.log("canot save visit", err)
		},
	})
}

function removeVisit(ev, i) {
	reqwest({
		url: apiserver + '/visits?id=' + propsCtl.point.id + '&vn=' + propsCtl.visited[i].idx,
		method: 'DELETE',
		crossOrigin: true,
		success: (data) => {
			propsCtl.dropVisit(i)
		},
		error: (err) => {
			console.log("canot remove visit", err)
		},
	})
}

//
// On-load
//

reqwest({
		url: apiserver + '/geos',
		method: 'GET',
		type: 'json',
		crossOrigin: true,
		success: (data) => {
			if (data.areas) {
				data.areas.forEach((item, i) => {
					item.state = "loading"
					areasLayer.addArea(item)
				})
			}
			if (data.points) {
				data.points.forEach((item, i) => {
					pointsLayer.addPoint(item)
				})
			}
		},
})
