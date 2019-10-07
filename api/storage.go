package main

type Geos interface {
	SavePoint(*SaveGeoReq) error
	LoadGeos() (*LoadGeosResp, error)
	RemoveGeo(int, string) (bool, error)

	SaveVisit(int, *SaveVisitReq) error
	LoadVisits(int) (*LoadVisitsResp, error)
	RemoveVisit(int, int) (bool, error)
}

func geos(c *Claims) Geos {
	return &LocalJsonGeos{
		name: c.UserId,
	}
}
