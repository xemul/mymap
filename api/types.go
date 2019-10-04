package main

type Area struct {
	Id		int			`json:"id"`
	Name		string			`json:"name"`
	Type		string			`json:"type"`
	Countries	[]string		`json:"countries"`
}

type Point struct {
	Id		int			`json:"id"`
	Name		string			`json:"name"`
	Lat		float64			`json:"lat"`
	Lng		float64			`json:"lng"`
	Area		int			`json:"area"`
	Countries	[]string		`json:"countries"`
}

type Visit struct {
	Date		string			`json:"date"`
	Tags		[]string		`json:"tags"`
	PId		*int			`json:"point,omitempty"`
}

type SavePointReq struct {
	Point		*Point			`json:"point,omitempty"`
	Areas		[]*Area			`json:"areas"`
}

type LoadGeosResp struct {
	Points		[]*Point		`json:"points"`
	Areas		[]*Area			`json:"areas"`
}

type SaveVisitReq struct {
	Visit					`json:",inline"`
}

type LoadVisitsResp struct {
	A		[]*Visit		`json:"array"`
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
	Visits		[]*Visit	`json:"visits,omitempty"`
}
