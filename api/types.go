package main

import (
	"encoding/json"
)

/*
 * Fundamental types
 */

type Map struct {
	Id		Id			`json:"id"`
	Name		string			`json:"name"`
	Copy		*int			`json:"copy,omitempty"`
}

func (m *Map)SetId(id Id) { m.Id = id }

type Area struct {
	Id		Id			`json:"id"`
	Name		string			`json:"name"`
	Type		string			`json:"type"`
	Countries	[]string		`json:"countries"`
}

func (a *Area)SetId(id Id) { a.Id = id }

type Point struct {
	Id		Id			`json:"id"`
	Name		string			`json:"name"`
	Lat		float64			`json:"lat"`
	Lng		float64			`json:"lng"`
	Area		Id			`json:"area"`
	Countries	[]string		`json:"countries"`
}

func (p *Point)SetId(id Id) { p.Id = id }

type Visit struct {
	Date		string			`json:"date"`
	Tags		[]string		`json:"tags"`
	Rating		int			`json:"rating"`
	PId		*Id			`json:"point,omitempty"`
}

type PointX struct {
	Point					`json:",inline"`
	Vis		[]*Visit		`json:"visits,omitempty"`
}

/*
 * API
 */

type LoadMapsResp struct {
	M		[]*Map			`json:"maps"`
}

type SaveGeoReq struct {
	Point		*Point			`json:"point,omitempty"`
	Areas		[]*Area			`json:"areas"`
}

type LoadGeosResp struct {
	Points		[]*Point		`json:"points"`
	Areas		[]*Area			`json:"areas"`
}

type LoadVisitsResp struct {
	A		[]*Visit		`json:"array"`
}

type RawGeos struct {
	Areas		json.RawMessage		`json:"areas"`
	Points		json.RawMessage		`json:"points"`
}
