package main

type Area struct {
	Id		int			`json:"id"`
	Name		string			`json:"name"`
	Type		string			`json:"type"`
}

type Point struct {
	Id		int			`json:"id"`
	Name		string			`json:"name"`
	Lat		float64			`json:"lat"`
	Lng		float64			`json:"lng"`
	Area		int			`json:"area"`
	Countries	[]string		`json:"countries"`
}

type SavePointReq struct {
	Point		*Point			`json:"point,omitempty"`
	Areas		[]*Area			`json:"areas"`
}

type LoadGeosResp struct {
	Points		[]*Point		`json:"points"`
	Areas		[]*Area			`json:"areas"`
}

type LocalJsonFile struct {
	Points		map[int]*PointFile	`json:"points"`
	Areas		map[int]*AreaFile	`json:"areas"`
}

type AreaFile struct {
	Area				`json:",inline"`
}

type PointFile struct {
	Point				`json:",inline"`
}
