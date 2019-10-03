package main

type Storage interface {
	SavePoint(*SavePointReq) error
	LoadGeos() (*LoadGeosResp, error)
	RemoveGeo(int, string) (bool, error)

	SaveVisit(int, *SaveVisitReq) error
	LoadVisits(int) (*LoadVisitsResp, error)
	RemoveVisit(int, int) (bool, error)
}

func storage(c *Claims) Storage {
	return &LocalJsonStorage{
		name: c.UserId,
	}
}
