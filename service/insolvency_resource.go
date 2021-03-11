package service

import (
	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/dao"
)

// InsolvencyResourceService contains the DAO for db access
type InsolvencyResourceService struct {
	DAO    dao.Service
	Config *config.Config
}
