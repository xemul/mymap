package main

type UDB interface {
	Close()

	openMDB(int, bool) (MDB, error)
	List() ([]*Map, error)
	Create(*Map) (error)
	Remove(int) (error)
}

type MDB interface {
	Id() int
	Close()

	Raw() ([]byte, error)

	SavePoint(*SaveGeoReq) error
	LoadGeos() (*LoadGeosResp, error)
	RemoveGeo(int, string) (bool, error)

	SaveVisit(int, *SaveVisitReq) error
	LoadVisits(int) (*LoadVisitsResp, error)
	RemoveVisit(int, int) (bool, error)
}

func openDB(c *Claims) UDB {
	return localUInfo(c.UserId)
}
