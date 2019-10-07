let config = {}

function highlightPoint(ev, pnt) {
	mymap.setView(pnt, highlightZoom)
	pnt.marker.setIcon(placeDIcon)
	setTimeout(() => { pnt.marker.setIcon(placeIcon) }, highlightTimeout)
}

function highlightArea(ev, area) {
	let c = area.layer.getBounds().getCenter()
	mymap.setView(c, highlightAZoom)
	area.layer.setStyle(areaHStyle)
	setTimeout(() => { area.layer.setStyle(areaStyle) }, highlightTimeout)
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
			mymap.addLayer(pointsLayer.lr)
		} else {
			mymap.removeLayer(pointsLayer.lr)
		}
	}
)

backendRq = function(rq) {
	if (!config.backend) {
		rq.warn("config load error")
		return
	}

	let headers = {}

	if (config.viewmap) {
		if (rq.method != "GET") {
			statusCtl.warn("View mode, data will not be saved")
			rq.success({})
			return
		}

		if (!rq.q) {
			rq.q = []
		}
		rq.q.push('viewmap=' + config.viewmap)
	} else {
		if (menuCtl.sess == null) {
			rq.error({message: "login not checked yet"})
			return
		}

		if (menuCtl.sess.user == null) {
			statusCtl.warn("You're not logged in, data will not be saved")
			rq.success({})
			return
		}

		headers = {
			Authorization: menuCtl.sess.user.token,
		}

		if (mapsCtl.current) {
			headers["X-MapId"] = mapsCtl.current
		}
	}

	let url = config.backend + rq.url

	if (rq.q) {
		url += '?' + rq.q.join('&')
	}

	axios({
		method: rq.method,
		url: url,
		data: rq.data,
		headers: headers,
	}).then((resp) => { rq.success(resp.data) }).catch((err) => { rq.error(err) })
}

Vue.component('rateimg', {
	props: ['rt'],
	template:
		`<span>
		<span v-if='rt == "-2"'><img src='static/img/awful.svg' title='awful'></span>
		<span v-if='rt == "-1"'><img src='static/img/bad.svg' title='bad'></span>
		<span v-if='rt == "0"'><img src='static/img/neutral.svg' title='neutral'></span>
		<span v-if='rt == "1"'><img src='static/img/good.svg' title='good'></span>
		<span v-if='rt == "2"'><img src='static/img/excellent.svg' title='excellent'></span>
		</span>`,

})

Vue.component('flagimg', {
	props: ['country'],
	template:
		`<span>
		<span v-if="country != ''">
		<img v-bind:src="'static/img/flags/' + country + '.svg'" class="flag" v-bind:alt='country' v-bind:title='country'>
		</span>
		<span v-if="country == ''">
		<img src="static/img/world.svg" title="click on a flag">
		</span>
		</span>`,
})

Vue.component('flaglist', {
	props: ['list'],
	template:
		`<span>
		<template v-for="f in list">
			<img v-bind:src="'static/img/flags/' + f + '.svg'" class="flag" v-bind:alt='f' v-bind:title='f'>
		</template>
		</span>`,
})

//
// Hidebar stuff
//

var hidebar = { n: "" }

hidebar.show = (w) => {
	if (hidebar.n == "") {
		mapCtl.resize("50%", mapHeight)
	}

	hidebar.n = w
}

hidebar.close = () => {
	hidebar.n = ""
	mapCtl.resize(mapWidth, mapHeight)
}

//
// Maps
//

var mapsCtl = new Vue({
	el: '#maps',
	data: {
		show: hidebar,
		maps: [],
		current: "",
		share: "",
		nMap: "",
	},
	methods: {
		closeMaps: () => { hidebar.close() },

		saveAreas: () => {},
		loadAreas: () => {},

		addMap: () => {
			backendRq({
				url: '/maps',
				method: 'post',
				data: JSON.stringify({name: mapsCtl.nMap}),
				success: (data) => {
					mapsCtl.nMap = ""
					mapsCtl.maps.push(data)
				},
				error: (err) => {
					statusCtl.err("Cannot create map: " + err.message)
				},
			})
		},

		removeMap: (ev, map) => {
		},

		switchMap: (ev, map) => {
			console.log("switch to map", map.id)
			clearMap()
			mapsCtl.current = map.id
			if (config.viewmap == "") {
				mapsCtl.share = "/map?viewmap=" + mapsCtl.current
			}
			loadGeos()
		},
	},
})

//
// Sidebar stuff
//

var menuCtl = new Vue({
	el: '#menu',
	data: {
		sess: null,
		viewmap: null,
	},
	methods: {
		showAreas: () => { hidebar.show("areas") },
		showPoints: () => { hidebar.show("points") },
		showMaps: () => { hidebar.show("maps") },
		showTimeline: () => {
			hidebar.show("timeline")
			timelineCtl.load()
		},
		showRating: () => {
			hidebar.show("ratings")
			ratingCtl.load()
		},
	},
})

var statusCtl = new Vue({
	el: '#status',
	data: {
		message: "",
		type: "",
		url: "",
	},
	methods: {
		err: (txt) => {
			console.log(txt)
			statusCtl.type = "error"
			statusCtl.message = txt
			setTimeout(() => { statusCtl.clear() }, errorTimeout)
		},

		warn: (txt) => {
			statusCtl.type = "warning"
			statusCtl.message = txt
			setTimeout(() => { statusCtl.clear() }, errorTimeout)
		},

		info: (txt) => {
			statusCtl.type = "info"
			statusCtl.message = txt
		},

		clear: () => {
			statusCtl.type = ""
			statusCtl.message = ""
		},
	},
})

function getCountries(area) {
	let ret = []
	if (area.countries) {
		area.countries.forEach((item, i) => { ret.push(item.code.toLowerCase()) })
	}
	return ret
}

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
					id:		area.id,
					name:		area.name,
					state:		"loading",
					type:		area.type,
					countries:	getCountries(area),
				})

				if (smallest == null || smallest.type < area.type) {
					smallest = area
				}
			})
			areas.sort((a,b) => { return strCmp(a.type, b.type) })

			let inside = { countries: [] }

			if (smallest) {
				inside.countries = getCountries(smallest)
				inside.area = smallest.id
			}

			selectionCtl.available = areas
			markerCtl.inside = inside
		},

		loadSelected: () => {
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

			backendRq({
				url: '/geos',
				method: 'POST',
				data: JSON.stringify(rq),
				success: (x) => {
					selectionCtl.commit(rq)
				},
				error: (err) => {
					statusCtl.err("Cannot save point: " + err.message)
				},
			})
		},

		commit: (rq) => {
			rq.areas.forEach((item, i) => { areasLayer.addArea(item) })
			if (rq.point) {
				pointsLayer.addPoint(rq.point)
			}

			markerLayer.remove()
			selectionCtl.clearSelection()
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

	axios.get('https://global.mapit.mysociety.org/point/4326/'+e.latlng.lng+','+e.latlng.lat).
		then((resp) => {
			selectionCtl.setAvailable(resp.data)
		}).
		catch((err) => {
			statusCtl.err("cannot find areas at the point: " + err.message)
			markerCtl.clearMarker()
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
	methods: {
		clearMarker: (ev) => {
			markerLayer.remove()
			selectionCtl.clearSelection()
		},
	},
})

//
// Areas
//

var areasLayer = areasLayer || {};
areasLayer.lr = L.featureGroup().addTo(mymap);

areasLayer.areaLoaded = function(area, data) {
	var lr = new L.GeoJSON(data, { style: areaStyle });
	lr.on('dblclick', function(e) {
	    var z = mymap.getZoom() + (e.originalEvent.shiftKey ? -1 : 1);
	    mymap.setZoomAround(e.containerPoint, z);
	});

	areasLayer.lr.addLayer(lr)
	areasCtl.updateArea(area, lr)
};

areasLayer.addArea = function(area) {
	if (!areasCtl.loaded[area.id]) {
		areasCtl.addArea(area)

		axios.get('https://global.mapit.mysociety.org/area/' + area.id + '.geojson?simplify_tolerance=0.0001').
			then((resp) => {
				areasLayer.areaLoaded(area, resp.data)
			}).
			catch((err) => {
				statusCtl.err("cannot load area: " + err.message)
			})
	}
}

sortAreas = new toggle(3)

var areasCtl = new Vue({
	el: '#areas',
	data: {
		loaded: {},
		nr: 0,
		show: hidebar,
		sorted: sortAreas,
		pc: "",
	},
	computed: {
		loadedS: () => {
			let ret = []
			Object.entries(areasCtl.loaded).forEach((a) => {
				let area = a[1]
				let found = true

				if (areasCtl.pc != "") {
					found = hasCountry(area.countries, areasCtl.pc)
				}

				if (found) {
					ret.push(a[1])
				}
			})
			if (areasCtl.sorted.i != 0) {
				ret.sort((a, b) => {
					let o = strCmp(a.name, b.name)
					if (areasCtl.sorted.i == 2) { o = -o }
					return o
				})
			}
			return ret
		},
	},
	methods: {
		clearAll: () => {
			areasCtl.loaded = {}
			areasCtl.nr = 0
			areasLayer.lr.clearLayers()
		},

		preferCountry: (ev, c) => {
		       areasCtl.pc = c
		},

		addArea: (area) => {
			areasCtl.$set(areasCtl.loaded, area.id, area)
			areasCtl.nr += 1
		},

		updateArea: (area, layer) => {
			area.state = "ready"
			area.layer = layer
			areasCtl.$set(areasCtl.loaded, area.id, area)
		},

		removeArea: (ev, area) => {
			backendRq({
					url: '/geos',
					q: [ 'type=area', 'id=' + area.id ],
					method: 'DELETE',
					success: (x) => {
						areasCtl.dropArea(area)
					},
					error: (err) => {
						statusCtl.err("Cannot remove area: " + err.message)
					},
			})
		},

		dropArea: (area) => {
			areasLayer.lr.removeLayer(L.stamp(area.layer))
			areasCtl.$delete(areasCtl.loaded, area.id)
			areasCtl.nr -= 1
		},

		closeAreas: () => { hidebar.close() },
	},
})

//
// Points
//

sortPointNames = new toggle(3)

var pointsLayer = pointsLayer || {}
pointsLayer.lr = L.layerGroup().addTo(mymap);

pointsLayer.addPoint = function(pt) {
	pt.marker = L.marker(pt, {icon: placeIcon}).addTo(pointsLayer.lr)
	pt.marker.bindTooltip(pt.name, {direction: "auto", opacity: placeTolltipOpacity})
	pt.marker.on('click', function(e) {
		propsCtl.showPoint(pt)
	})
	pointsCtl.addPoint(pt)
}

var pointsCtl = new Vue({
	el: '#points',
	data: {
		loaded: {},
		nr: 0,
		hide: hidePointsToggle,
		show: hidebar,
		sortedN: sortPointNames,
	},
	computed: {
		loadedS: () => {
			let buckets = {}

			Object.entries(pointsCtl.loaded).forEach((a) => {
				let pt = a[1]
				let bkey = pt.countries.join(',')

				if (!buckets[bkey]) {
					buckets[bkey] = {
						countries: pt.countries,
						pts: [],
					}
				}

				buckets[bkey].pts.push(pt)
			})

			let ret = []

			Object.entries(buckets).forEach((e) => {
				let bkt = e[1]
				if (sortPointNames.i != 0) {
					bkt.pts.sort((a,b) => {
						let o = strCmp(a.name, b.name)
						if (sortPointNames.i == 2) { o = -o }
						return o
					})
				}
				ret.push(bkt)
			})

			ret.sort((a,b) => { return b.pts.length - a.pts.length })

			return ret
		},
	},
	methods: {
		clearAll: () => {
			pointsCtl.loaded = {}
			pointsCtl.nr = 0
			pointsLayer.lr.clearLayers()
		},

		addPoint: (pt) => {
			pointsCtl.$set(pointsCtl.loaded, pt.id, pt)
			pointsCtl.nr += 1
		},

		removePoint: (ev, pnt) => {
			backendRq({
					url: '/geos',
					q: [ 'type=point', 'id=' + pnt.id ],
					method: 'DELETE',
					success: (x) => {
						pointsCtl.dropPoint(pnt)
					},
					error: (err) => {
						statusCtl.err("Cannot remove point: " + err.message)
					},
			})
		},

		dropPoint: (pnt) => {
			pointsLayer.lr.removeLayer(pnt.marker)
			pointsCtl.$delete(pointsCtl.loaded, pnt.id)
			pointsCtl.nr -= 1
		},

		closePoints: () => { hidebar.close() },
	},
})

//
// Ratings
//

var ratingCtl = new Vue({
	el: '#ratings',
	data: {
		state:		"",
		points:		[],
		show:		hidebar,
		sc:		"",
	},
	computed: {
		pointsS: () => {
			if (ratingCtl.sc == "") {
				return ratingCtl.points
			}

			let vf = []
			ratingCtl.points.forEach((v, i) => {
				if (hasCountry(v.pt.countries, ratingCtl.sc)) {
					vf.push(v)
				}
			})

			return vf
		},
	},

	methods: {
		selectCountry: (ev, c) => {
		       ratingCtl.sc = c
		},

		load: () => {
			ratingCtl.state = "loading"
			ratingCtl.points = []

			backendRq({
				url: '/visits',
				method: 'GET',
				success: (data) => {
					ratingCtl.state = "ready"
					if (data.array) {
						let points = {}

						data.array.forEach((v, i) => {
							let pt = points[v.point]
							if (pt == null) {
								pt = {
									pt: pointsCtl.loaded[v.point],
									rating: v.rating,
									nr: 1,
								}

								points[v.point] = pt
							} else {
								pt.rating += v.rating
								pt.nr += 1
							}
						})

						Object.keys(points).forEach((key, idx) => {
							let pt = points[key]
							pt.rating /= pt.nr
							pt.irating = Math.round(pt.rating)
							ratingCtl.points.push(pt)
						})

						ratingCtl.points.sort((a, b) => { return b.rating - a.rating })
					}
				},
				error: (err) => {
					statusCtl.err("Cannot load visits: " + err.message)
				},
			})
		},

		closeRatings: () => {
			ratingCtl.state = ""
			ratingCtl.points = []
			hidebar.close()
		},
	},
})

//
// Timeline
//

var timelineCtl = new Vue({
	el: '#timeline',
	data: {
		state:		"",
		visited:	[],
		show:		hidebar,
		sc:		"",
	},
	computed: {
		visitedS: () => {
			if (timelineCtl.sc == "") {
				return timelineCtl.visited
			}

			let vf = []
			timelineCtl.visited.forEach((v, i) => {
				if (hasCountry(v.pt.countries, timelineCtl.sc)) {
					vf.push(v)
				}
			})

			return vf
		},
	},

	methods: {
		selectCountry: (ev, c) => {
		       timelineCtl.sc = c
		},

		sortVisited: () => {
			timelineCtl.visited.sort((a,b) => {
				return dateScore(b.date) - dateScore(a.date)
			})
		},

		load: () => {
			timelineCtl.state = "loading"
			timelineCtl.visited = []

			backendRq({
				url: '/visits',
				method: 'GET',
				success: (data) => {
					timelineCtl.state = "ready"
					if (data.array) {
						data.array.forEach((v, i) => {
							v.pt = pointsCtl.loaded[v.point]
							v.idx = i
							timelineCtl.visited.push(v)
						})
						timelineCtl.sortVisited()
					}
				},
				error: (err) => {
					statusCtl.err("Cannot load visits: " + err.message)
				},
			})
		},

		closeTimeline: () => {
			timelineCtl.state = ""
			timelineCtl.visited = []
			hidebar.close()
		},
	},
})

//
// Props
//

var propsCtl = new Vue({
	el: '#pprops',
	data: {
		point: null,
		visited: [],
		nvDate: "",
		nvTags: "",
		nvRate: 0,
		visits: false,
	},
	methods: {
		closeProps: () => {
			propsCtl.point = null
			propsCtl.visited = []
			propsCtl.clearNv()
		},

		clearNv: () => {
			propsCtl.nvDate = ""
			propsCtl.nvTags = ""
			propsCtl.nvRate = 0
		},

		showPoint: (pt) => {
			propsCtl.point = pt
			propsCtl.visited = []
			propsCtl.clearNv()

			backendRq({
				url: '/visits',
				q: [ 'id=' + pt.id ],
				method: 'GET',
				success: (data) => {
					if (data.array) {
						data.array.forEach((v, i) => {
							v.idx = i
							propsCtl.visited.push(v)
						})
						propsCtl.sortVisited()
					}
				},
				error: (err) => {
					statusCtl.err("Cannot load visits: " + err.message)
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

		addVisit: () => {
			let nv = {
				date: propsCtl.nvDate,
				tags: propsCtl.nvTags.split(/\s*,\s*/),
				rating: parseInt(propsCtl.nvRate),
			}

			backendRq({
				url: '/visits',
				q: [ 'id=' + propsCtl.point.id ],
				method: 'POST',
				contentType: 'application/json',
				data: JSON.stringify(nv),
				success: (data) => {
					propsCtl.commit(nv)
				},
				error: (err) => {
					statusCtl.err("Cannot save visit: " + err.message)
				},
			})
		},

		removeVisit: (ev, i) => {
			backendRq({
				url: '/visits',
				q: [ 'id=' + propsCtl.point.id, 'vn=' + propsCtl.visited[i].idx ],
				method: 'DELETE',
				success: (data) => {
					propsCtl.dropVisit(i)
				},
				error: (err) => {
					statusCtl.err("Cannot remove visit: " + err.message)
				},
			})
		},
	},
})

function dateScore(date) {
	let ds = date.split("/")

	let y = parseInt(ds.pop()) || 0
	let m = parseInt(ds.pop()) || 0
	let d = parseInt(ds.pop()) || 0

	return ((y * 12) + m) * 32 + d
}

function strCmp(a, b) {
	if (a > b) {
		return 1
	} else if (a < b) {
		return -1
	} else {
		return 0
	}
}

function hasCountry(cl, c) {
	let found = false
	cl.forEach((x) => { if (c == x) { found = true } })
	return found
}

//
// On-load
//

axios.get('/config')
	.then((resp) => {
		config = resp.data
		console.log("config: ", config)
		if (config.viewmap) {
			menuCtl.viewmap = config.viewmap
		}
		login()
	})
	.catch((err) => { statusCtl.warn("failed to load config") })

function login() {
	if (config.viewmap) {
		menuCtl.sess = { user: null }
		loadGeos()
		return
	}

	axios.get('/creds').
		then((resp) => {
			console.log("authorized as ", resp.data.id)
			menuCtl.sess = { user: resp.data }
			propsCtl.visits = true
			loadMaps()
		}).
		catch((err) => {
			console.log("anonymous mode")
			menuCtl.sess = { user: null }
		})
}

function loadMaps() {
	backendRq({
			url: '/maps',
			method: 'GET',
			success: (data) => {
				mapsCtl.maps = data.maps
				mapsCtl.current = data.maps[0].id
				console.log("select map: ", mapsCtl.current)
				if (config.viewmap == "") {
					mapsCtl.share = "/map?viewmap=" + mapsCtl.current
				}
				loadGeos()
			},
			error: (err) => {
				statusCtl.err("Cannot load maps: " + err.message)
			},
	})
}

function clearMap() {
	areasCtl.clearAll()
	pointsCtl.clearAll()
}

function loadGeos() {
	backendRq({
			url: '/geos',
			method: 'GET',
			success: (data) => {
				console.log("loaded geos ", data.areas, data.points)
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
			error: (err) => {
				statusCtl.err("Cannot load points and areas: " + err.message)
			},
	})
}
