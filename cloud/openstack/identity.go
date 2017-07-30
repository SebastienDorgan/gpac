package openstack

import "time"

type Service struct {
	Id     string
	Name   string
	Region string
	Url    string
}

type Catalog map[string]Service

type AccessData struct {
	AuthToken string
	TokenExpiration time.Time
	Catalog   Catalog
}

type AuthAPI interface {
	AuthRequest() string
	Authenticate() (AccessData, error)
}



