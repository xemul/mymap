package main

type Geos interface {
	SavePoint(*SaveGeoReq) error
	LoadGeos() (*LoadGeosResp, error)
	RemoveGeo(int, string) (bool, error)

	SaveVisit(int, *SaveVisitReq) error
	LoadVisits(int) (*LoadVisitsResp, error)
	RemoveVisit(int, int) (bool, error)
}

func openMap(uid string, mapid int) (Geos, error) {
	return OpenLocalMap(uid, mapid)
}

func listMaps(uid string) ([]*Map, error) {
	return ListLocalMaps(uid)
}

func createMap(uid string, m *Map) error {
	return CreateLocalMap(uid, m)
}
