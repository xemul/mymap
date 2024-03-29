let config = {}

function highlightPoint(pnt) {
	mymap.setView(pnt, highlightZoom)
	pnt.marker.setIcon(placeDIcon)
	setTimeout(() => { pnt.marker.setIcon(placeIcon) }, highlightTimeout)
}

function sizeZoom(sz) {
	if (sz > 50) {
		return 3
	}
	if (sz > 25) {
		return 4
	}
	if (sz > 12) {
		return 5
	}
	if (sz > 6) {
		return 6
	}
	if (sz > 3) {
		return 7
	}
	if (sz > 1) {
		return 8
	}
	if (sz > 0.5) {
		return 9
	}
	if (sz > 0.25) {
		return 10
	}
	if (sz > 0.125) {
		return 11
	}

	return 12
}

function getZoom(b) {
	let ne = b.getNorthEast()
	let sw = b.getSouthWest()
	let sz = Math.max(Math.abs(sw.lat - ne.lat), Math.abs(sw.lng - ne.lng))
	return sizeZoom(sz)
}

function highlightArea(area) {
	if (!area.layer) {
		return
	}

	let b = area.layer.getBounds()
	let c = b.getCenter()
	let z = getZoom(b)
	mymap.setView(c, z)
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

	if (menuCtl.sess == null) {
		rq.error({message: "login not checked yet"})
		return
	}

	let headers = {}

	if (menuCtl.sess.user != null) {
		headers = {
			Authorization: menuCtl.sess.user.token,
		}
	} else if (rq.method != "get") {
		statusCtl.warn("You're not logged in, data will not be saved")
		rq.success({})
		return
	}

	axios({
		method: rq.method,
		url: config.backend + rq.url,
		data: rq.data,
		headers: headers,
	})
	.then((resp) => { rq.success(resp.data) })
	.catch((err) => {
		if (err.response.status == 401) {
			statusCtl.err("Token expired, please, re-login")
			menuCtl.sess = { user: null }
		} else {
			rq.error(err)
		}
	})
}

Vue.component('pointln', {
	props: ['pnt'],
	template:
		`<span>
		<a href="#" v-on:click=highlightPoint(pnt)><img src="static/img/locate.svg"></a>
		{{pnt.name}}
		</span>`,
})

Vue.component('arealn', {
	props: ['area'],
	template:
		`<span>
		<a href="#" v-on:click=highlightArea(area)><img src="static/img/locate.svg"></a>
		{{area.name}}
		</span>`,
})

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
// Sidebar
//

var sidebarSwitch = {
	n: "",
	clear: null,
}

sidebarSwitch.show = (name, clear) => {
	if (sidebarSwitch.clear != null) {
		sidebarSwitch.clear()
	}

	sidebarSwitch.n = name
	sidebarSwitch.clear = clear
}

sidebarSwitch.close = () => {
	if (sidebarSwitch.clear != null) {
		sidebarSwitch.clear()

		sidebarSwitch.n = ""
		sidebarSwitch.clear = null
	}
}

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
		sidebar: sidebarSwitch,

		maps: {},
		current: null,
		share: "",
		nMap: "",
		copyMap: false,
		mName: "",
	},
	computed: {
		mapsS: () => {
			let ret = [ ]
			let defmap = null

			Object.entries(mapsCtl.maps).forEach((x) => {
				let mp = x[1]

				if (mp.name == "default") {
					defmap = mp
				} else {
					ret.push(mp)
				}
			})

			ret.sort((a, b) => { return strCmp(a.name, b.name) })
			if (defmap) {
				ret.unshift(defmap)
			}

			return ret
		},

		stats: () => {
			let countries = {}

			Object.entries(areasCtl.loaded).forEach((a) => {
				let area = a[1]
				area.countries.forEach((cn, i) => { countries[cn] = true })
			})

			return {
				countries: Object.keys(countries).length,
				areas: areasCtl.nr,
				points: pointsCtl.nr,
			}
		},
	},
	methods: {
		editMap: () => { mapsCtl.mName = mapsCtl.current.name },
		unsaveMap: () => { mapsCtl.mName = "" },
		saveMap: () => {
			if (mapsCtl.mName != mapsCtl.current.name) {
				backendRq({
					url: '/maps/' + mapsCtl.current.id,
					method: 'patch',
					data: { name: mapsCtl.mName },
					success: (data) => {
						mapsCtl.current.name = mapsCtl.mName
						mapsCtl.mName = ""
					},
					error: (err) => {
						statusCtl.err("Cannot save map name: ", err.message)
					},
				})
			} else {
				mapsCtl.mName = ""
			}
		},

		closeMaps: () => { sidebarSwitch.close() },
		clearMaps: () => {},

		saveMapGeos: () => {
			console.log("Will save map geos")
			window.open(config.backend + "/maps/" + mapsCtl.current.id)
		},

		setMaps: (lst) => {
			lst.forEach((m, i) => {
				mapsCtl.maps[m.id] = m
			})
		},

		addMap: () => {
			console.log("copy: ", mapsCtl.copyMap)

			let q = {name: mapsCtl.nMap}
			if (mapsCtl.copyMap) {
				q["copy"] = mapsCtl.current.id
			}

			backendRq({
				url: '/maps',
				method: 'post',
				data: q,
				success: (data) => {
					mapsCtl.nMap = ""
					mapsCtl.copyMap = false
					mapsCtl.$set(mapsCtl.maps, data.id, data)
				},
				error: (err) => {
					statusCtl.err("Cannot create map: " + err.message)
				},
			})
		},

		removeMap: (ev, map) => {
			backendRq({
				url: '/maps/' + map.id,
				method: 'delete',
				success: (data) => {
					mapsCtl.$delete(mapsCtl.maps, map.id)
				},
				error: (err) => {
					statusCtl.err("Cannot remove map: " + err.message)
				},
			})
		},

		switchMap: (ev, map) => {
			console.log("switch to map", map.id)
			clearMap()
			mapsCtl.current = map
			menuCtl.current = map.name
			if (config.viewmap == "") {
				mapsCtl.share = "/map?viewmap=" + map.id
			}
			loadGeos()
		},
	},
})

var finderCtl = new Vue({
	el: '#finder',
	data: {
		sidebar: sidebarSwitch,
		findName: "",
		places: [],
	},
	methods: {
		closeFinder: () => { sidebarSwitch.close() },
		clearSearch: () => { },

		locatePlace: (ev, plc) => {
			mymap.setView(plc.loc, sizeZoom(plc.size))
			let c = L.circle(plc.loc, {radius: 10000 * plc.size}).addTo(mymap)
			setTimeout(() => { c.remove() }, highlightTimeout)
		},

		searchName: () => {
			finderCtl.places = []
			axios.get('https://nominatim.openstreetmap.org/search?q=' + finderCtl.findName + '&format=json')
				.then((resp) => {
					resp.data.forEach((el, i) => {
						finderCtl.places.push({
							name: el.display_name,
							type: el.type,
							loc: {lat: el.lat, lng: el.lon},
							size: Math.max(Math.abs(el.boundingbox[0] - el.boundingbox[1]),
									Math.abs(el.boundingbox[2] - el.boundingbox[3])),
						})
					})
				})
				.catch((err) => {
					statusCtl.err("Cannot search by name: " + err.message)
				})
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
		current: "",
	},
	methods: {
		showFinder: () => {
			sidebarSwitch.show("finder", finderCtl.clearSearch)
		},

		showMaps: () => {
			sidebarSwitch.show("maps", mapsCtl.clearMaps)
		},

		showPoints: () => {
			sidebarSwitch.show("points", pointsCtl.clearPoints)
		},

		showAreas: () => {
			sidebarSwitch.show("areas", areasCtl.clearAreas)
		},

		showCountries: () => {
			sidebarSwitch.show("countries", countriesCtl.clearCountries)
		},

		showTimeline: () => {
			sidebarSwitch.show("timeline", timelineCtl.clearTimeline)
			timelineCtl.load()
		},
		showRating: () => {
			sidebarSwitch.show("ratings", ratingCtl.clearRatings)
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
	if (area.codes) {
		let code = area.codes["iso3166_1"]
		if (code) {
			return [ code.toLowerCase() ]
		}
	}

	let ret = []
	if (area.countries) {
		area.countries.forEach((item, i) => { ret.push(item.code.toLowerCase()) })
	}

	if (ret.length == 0) {
		return [ "xx" ]
	}

	return ret
}

var selectionCtl = new Vue({
	el: '#selection',
	data: {
		sidebar: sidebarSwitch,

		available: [],
		selected: [],
		pointName: "",

		show: showTypesToggle,
	},
	methods: {
		clearSelection: () => {
			selectionCtl.pointName = ""
			selectionCtl.available = []
			selectionCtl.selected = []
		},

		setAvailable: (data) => {
			let areas = parseAreas(data)
			let inside = getInside(areas)

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
				url: '/maps/' + mapsCtl.current.id + '/geos',
				method: 'post',
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

			markerCtl.closeMarker()
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
					highlightPoint(pt)
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
	markerCtl.showMarker(e.latlng, null)
	selectionCtl.clearSelection()

	axios.get('https://global.mapit.mysociety.org/point/4326/'+e.latlng.lng+','+e.latlng.lat).
		then((resp) => {
			selectionCtl.setAvailable(resp.data)
		}).
		catch((err) => {
			statusCtl.err("cannot find areas at the point: " + err.message)
			markerCtl.closeMarker()
		})
})

function parseAreas(data) {
	let areas = []

	Object.keys(data).forEach((key, idx) => {
		let area = data[key]
		let a = {
			id:		area.id,
			name:		area.name,
			state:		"loading",
			type:		area.type,
			countries:	getCountries(area),
		}

		areas.push(a)
	})
	areas.sort((a,b) => { return strCmp(a.type, b.type) })

	return areas
}

function getInside(areas) {
	let inside = { countries: [] }

	if (areas.length > 0) {
		let smallest = areas[areas.length-1]
		inside.countries = smallest.countries
		inside.area = smallest.id
	}

	return inside
}

//
// Marker
//

var markerLayer = markerLayer || {}
markerLayer.lm = null

markerLayer.setMark = function(latlng) {
	if (markerLayer.lm != null) {
		markerLayer.lm.remove()
	}

	if (latlng) {
		markerLayer.lm = L.marker(latlng, {icon: pointIcon}).addTo(mymap);
	} else {
		markerLayer.lm = null
	}
}

var markerCtl = new Vue({
	el: '#marker',
	data: {
		sidebar: sidebarSwitch,

		latlng: null,
		inside: null,
	},
	methods: {
		showMarker: (latlng, ins) => {
			sidebarSwitch.show("marker", markerCtl.clearMarker)

			markerCtl.latlng = latlng
			markerCtl.inside = ins
			markerLayer.setMark(latlng)
		},

		closeMarker: () => {
			sidebarSwitch.close()
		},

		clearMarker: (ev) => {
			selectionCtl.clearSelection()

			markerCtl.latlng = null
			markerCtl.inside = null
			markerLayer.setMark(null)
		},
	},
})

var countriesCtl = new Vue({
	el: '#countries',
	data: {
		sidebar: sidebarSwitch,
	},
	computed: {
		countries: () => {
			let dict = {}

			Object.entries(areasCtl.loaded).forEach((a) => {
				let area = a[1]
				area.countries.forEach((cn, i) => { dict[cn] = true })
			})

			let list = []

			Object.entries(dict).forEach((a) => { list.push(a[0]) })

			list.sort()
			return list
		},
	},
	methods: {
		closeCountries: () => { sidebarSwitch.close() },
		clearCountries: () => {},
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
		sidebar: sidebarSwitch,

		loaded: {},
		nr: 0,
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
					url: '/maps/' + mapsCtl.current.id + '/geos/areas/' + area.id,
					method: 'delete',
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

		closeAreas: () => { sidebarSwitch.close() },
		clearAreas: () => {},
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
	pt.marker.on('dragend', function(e) {
		propsCtl.draggedPoint()
	})
	pointsCtl.addPoint(pt)
}

var pointsCtl = new Vue({
	el: '#points',
	data: {
		sidebar: sidebarSwitch,

		loaded: {},
		nr: 0,
		hide: hidePointsToggle,
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
					url: '/maps/' + mapsCtl.current.id + '/geos/points/' + pnt.id,
					method: 'delete',
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

		closePoints: () => { sidebarSwitch.close() },
		clearPoints: () => {},
	},
})

//
// Ratings
//

var ratingCtl = new Vue({
	el: '#ratings',
	data: {
		sidebar:	sidebarSwitch,

		state:		"",
		points:		[],
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
				url: '/maps/' + mapsCtl.current.id + '/visits',
				method: 'get',
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

		closeRatings: () => { sidebarSwitch.close() },

		clearRatings: () => {
			ratingCtl.state = ""
			ratingCtl.points = []
		},
	},
})

//
// Timeline
//

var timelineCtl = new Vue({
	el: '#timeline',
	data: {
		sidebar:	sidebarSwitch,

		state:		"",
		visited:	[],
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
				url: '/maps/' + mapsCtl.current.id + '/visits',
				method: 'get',
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

		closeTimeline: () => { sidebarSwitch.close() },

		clearTimeline: () => {
			timelineCtl.state = ""
			timelineCtl.visited = []
		},
	},
})

//
// Props
//

var propsCtl = new Vue({
	el: '#pprops',
	data: {
		sidebar: sidebarSwitch,

		point: null,
		visited: [],
		nvDate: "",
		nvTags: "",
		nvRate: 0,
		editable: false,

		editing: false,
		ptName: "",
		ptLatlng: null,
		ptInside: null
	},
	methods: {
		draggedPoint: () => {
			if (propsCtl.editing) {
				let newll = propsCtl.point.marker.getLatLng()
				propsCtl.ptLatlng = newll
				propsCtl.ptInside = null

				axios.get('https://global.mapit.mysociety.org/point/4326/'+newll.lng+','+newll.lat).
					then((resp) => {
						if (propsCtl.editing) {
							let areas = parseAreas(resp.data)
							let inside = getInside(areas)
							propsCtl.ptInside = inside
						}
					}).
					catch((err) => {
						statusCtl.err("cannot find areas at the new point: " + err.message)
					})
			}
		},

		editPoint: () => {
			propsCtl.ptName = propsCtl.point.name
			propsCtl.point.marker.setIcon(placeMIcon)
			propsCtl.point.marker.dragging.enable()

			propsCtl.editing = true
		},
		unsavePoint: () => {
			propsCtl.editing = false

			propsCtl.ptName = ""

			propsCtl.point.marker.setLatLng(propsCtl.point)
			propsCtl.point.marker.setIcon(placeIcon)
			propsCtl.point.marker.dragging.disable()
			propsCtl.ptLatlng = null
			propsCtl.ptInside = null
		},
		savePoint: () => {
			if (propsCtl.ptLatlng != null && propsCtl.ptInside == null) {
				statusCtl.warn("wait a bit...")
				return
			}

			let dirty = false
			let npt = {}

			if (propsCtl.ptName != propsCtl.point.name) {
				npt.name = propsCtl.ptName
				dirty = true
			}
			if (propsCtl.ptLatlng != null) {
				npt.lat = propsCtl.ptLatlng.lat
				npt.lng = propsCtl.ptLatlng.lng
				npt.area = propsCtl.ptInside.area
				npt.countries = propsCtl.ptInside.countries
				dirty = true
			}

			if (dirty) {
				backendRq({
					url: propsCtl.pntURL(),
					method: 'patch',
					data: npt,
					success: (data) => {
						if (npt.name) {
							propsCtl.point.name = npt.name
						}
						if (npt.area) {
							propsCtl.point.lat = npt.lat
							propsCtl.point.lng = npt.lng
							propsCtl.point.area = npt.area
							propsCtl.point.countries = npt.countries
						}

						propsCtl.unsavePoint()
					},
					error: (err) => {
						statusCtl.err("Cannot save point name: ", err.message)
					},
				})
			} else {
				propsCtl.unsavePoint()
			}
		},

		closeProps: () => { sidebarSwitch.close() },

		clearProps: () => {
			propsCtl.visited = []
			propsCtl.clearNv()
			propsCtl.unsavePoint()
			propsCtl.point = null
		},

		clearNv: () => {
			propsCtl.nvDate = ""
			propsCtl.nvTags = ""
			propsCtl.nvRate = 0
		},

		showPoint: (pt) => {
			sidebarSwitch.show("props", propsCtl.clearProps)

			propsCtl.point = pt
			propsCtl.visited = []
			propsCtl.clearNv()

			backendRq({
				url: '/maps/' + mapsCtl.current.id + '/geos/points/' + pt.id + '/visits',
				method: 'get',
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
				url: propsCtl.pntURL() + '/visits',
				method: 'post',
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
				url: propsCtl.pntURL() + '/visits/' + propsCtl.visited[i].idx,
				method: 'delete',
				success: (data) => {
					propsCtl.dropVisit(i)
				},
				error: (err) => {
					statusCtl.err("Cannot remove visit: " + err.message)
				},
			})
		},

		pntURL: () => {
			return  '/maps/' + mapsCtl.current.id + '/geos/points/' + propsCtl.point.id
		},

	},
})

function showPoint(ev, pnt) {
	highlightPoint(pnt)
	propsCtl.showPoint(pnt)
}

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
		mapsCtl.switchMap(null, { id: config.viewmap, name: 'shared' })
		return
	}

	axios.get('/creds').
		then((resp) => {
			console.log("authorized as ", resp.data.id)
			menuCtl.sess = { user: resp.data }
			propsCtl.editable = true
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
			method: 'get',
			success: (data) => {
				mapsCtl.setMaps(data.maps)
				let defid = 0
				data.maps.forEach((mp, idx) => {
					if (mp.name == 'default') { defid = idx }
				})
				mapsCtl.switchMap(null, data.maps[defid])
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
			url: '/maps/' + mapsCtl.current.id + '/geos',
			method: 'get',
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
			error: (err) => {
				statusCtl.err("Cannot load points and areas: " + err.message)
			},
	})
}

function uploadMapGeos(files) {
	backendRq({
		url: '/maps/' + mapsCtl.current.id,
		method: 'put',
		data: files[0],
		success: (data) => {
			statusCtl.info("Uploaded, please, reload the page")
		},
		error: (err) => {
			statusCtl.err("Cannot upload the map file")
		},
	})
}
