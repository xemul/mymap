const startAt = [ 55.0000, 40.0000 ]
const areaColor = "#4041ff"
const placeTolltipOpacity = 0.7

const pointIcon = L.icon({
	iconUrl:	'static/img/point.svg',
	iconSize:	[20, 30],
	iconAnchor:	[10, 29],
})

const placeIcon = L.icon({
	iconUrl:	'static/img/place.svg',
	iconSize:	[14, 21],
	iconAnchor:	[ 7, 20],
})

const placeDIcon = L.icon({
	iconUrl:	'static/img/place-d.svg',
	iconSize:	[16, 24],
	iconAnchor:	[ 8, 23],
})

const errorTimeout = 3000
const highlightTimeout = 1000
const highlightZoom = 7

const mapWidth = "75%"
const mapHeight = "98%"
