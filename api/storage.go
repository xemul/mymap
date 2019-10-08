package main

type UInfo interface {
	Close()

	Geos(int) (Geos, error)
	List() ([]*Map, error)
	Create(*Map) (error)
	Remove(int) (error)
}

type Geos interface {
	Close()

	SavePoint(*SaveGeoReq) error
	LoadGeos() (*LoadGeosResp, error)
	RemoveGeo(int, string) (bool, error)

	SaveVisit(int, *SaveVisitReq) error
	LoadVisits(int) (*LoadVisitsResp, error)
	RemoveVisit(int, int) (bool, error)
}

func openDB(c *Claims) UInfo {
	return localUInfo(c.UserId)
}
