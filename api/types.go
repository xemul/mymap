package main

type Area struct {
	Id	int		`json:"id"`
	Name	string		`json:"name"`
	Type	string		`json:"type"`
}

type SaveReq struct {
	Lat	float64		`json:"lat"`
	Lng	float64		`json:"lng"`
	Areas	[]*Area		`json:"areas"`
}

type LoadResp struct {
	Areas	[]*Area		`json:"areas"`
}

type LocalJsonFile struct {
	Areas	map[int]*AreaFile	`json:"areas"`
}

type AreaFile struct {
	Area			`json:",inline"`
}
