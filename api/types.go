package main

type Area struct {
	Id	int			`json:"id"`
	Name	string			`json:"name"`
	Type	string			`json:"type"`
}

type Point struct {
	Lat	float64			`json:"lat"`
	Lng	float64			`json:"lng"`
}

type SaveReq struct {
	Point				`json:",inline"`
	Areas	[]*Area			`json:"areas"`
}

type LoadResp struct {
	Points	[]*Point		`json:"points"`
	Areas	[]*Area			`json:"areas"`
}

type LocalJsonFile struct {
	Points	[]*PointFile		`json:"points"`
	Areas	map[int]*AreaFile	`json:"areas"`
}

type AreaFile struct {
	Area				`json:",inline"`
}

type PointFile struct {
	Point				`json:",inline"`
}
