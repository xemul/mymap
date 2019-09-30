var apiserver = "http://localhost:8082"

const startAt = [ 55.0000, 40.0000 ]
const areaColor = "#4041ff"
const placeTolltipOpacity = 0.7

const pointIcon = L.icon({
	iconUrl:	'img/point.svg',
	iconSize:	[28, 42],
	iconAnchor:	[14, 41],
})

const placeIcon = L.icon({
	iconUrl:	'img/place.svg',
	iconSize:	[14, 21],
	iconAnchor:	[ 7, 20],
})

const placeDIcon = L.icon({
	iconUrl:	'img/place-d.svg',
	iconSize:	[16, 24],
	iconAnchor:	[ 8, 23],
})

const highlightTimeout = 1000
const highlightZoom = 7
