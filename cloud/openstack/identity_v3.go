package openstack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type AuthAPIV3 struct {
	URL               string
	UserTokenId       *string
	UserName          *string
	UserId            *string
	UserPassword      *string
	UserDomainName    *string
	UserDomainId      *string
	ProjectDomainName *string
	ProjectDomainId   *string
	ProjectName       *string
	ProjectId         *string
	ExplicitlyUnScope bool
}

func fmtQualifier(id *string, name *string, domain *string) *string {

	if domain == nil {
		if id != nil {
			elt := fmt.Sprintf(`{ "id": "%s" }`, *id)
			return &elt
		} else if name != nil {
			elt := fmt.Sprintf(`{ "name": "%s" }`, *name)
			return &elt
		}
	} else {
		if id != nil {
			elt := fmt.Sprintf(`{ "id": "%s", "domain": %s }`, *id, *domain)
			return &elt
		} else if name != nil {
			elt := fmt.Sprintf(`{ "name": "%s", "domain": %s }`, *name, *domain)
			return &elt
		}
	}

	return nil
}

func fmtUser(id *string, name *string, password *string, domain *string) string {

	if domain == nil {
		if id != nil {
			return fmt.Sprintf(`{ "id": "%s", "password": "%s" }`, *id, *password)
		} else if name != nil {
			return fmt.Sprintf(`{ "name": "%s", "password": "%s" }`, *name, *password)
		}
	} else {
		if id != nil {
			return fmt.Sprintf(`{ "id": "%s", "password": "%s", "domain": %s }`, *id, *password, *domain)
		} else if name != nil {
			return fmt.Sprintf(`{ "name": "%s", "password": "%s", "domain": %s }`, *name, *password, *domain)
		}
	}
	return "null"
}

func (auth AuthAPIV3) authRequestWithPassword() string {
	request := `
	{
		"auth": {
			"identity": {
				"methods": ["password"],
				"password": {
					"user": %s
				}
			}
			%s
		}
	}
	`

	userDomain := fmtQualifier(auth.UserDomainId, auth.UserDomainName, nil)
	user := fmtUser(auth.UserId, auth.UserName, auth.UserPassword, userDomain)

	scope := auth.fmtScope()
	request = fmt.Sprintf(request, user, scope)
	return request
}

func (auth AuthAPIV3) authRequestWithToken() string {
	request := `
	{
		"auth": {
			"identity": {
				"methods": ["token"],
				"token": {
					"id": "%s"
				}
			}
			%s
		}
	}
	`

	scope := auth.fmtScope()
	request = fmt.Sprintf(request, auth.UserTokenId, scope)
	return request
}
func (auth AuthAPIV3) fmtScope() string {
	scope := ""
	projectDomain := fmtQualifier(auth.ProjectDomainId, auth.ProjectDomainName, nil)
	project := fmtQualifier(auth.ProjectId, auth.ProjectName, projectDomain)
	if project != nil {
		scope = `
			,
			"scope": {
				"project": %s
			}
			`
		scope = fmt.Sprintf(scope, *project)
	} else if auth.ExplicitlyUnScope {
		scope = `
			,
			"scope": "unscoped"
			`
	}
	return scope
}

func (auth AuthAPIV3) AuthRequest() string {
	if auth.UserTokenId == nil {
		return auth.authRequestWithPassword()
	} else {
		return auth.authRequestWithToken()
	}
}

type tokenObject map[string]interface{}
type endPointObject map[string]string

func (auth AuthAPIV3) Authenticate() (*AccessData, error) {
	resp, err := http.Post(auth.URL, "application/json", strings.NewReader(auth.AuthRequest()))
	if err != nil {
		return nil, err
	}
	access := AccessData{}
	access.AuthToken = resp.Header.Get("X-Subject-Token")

	var body map[string]tokenObject
	var bodyBuffer []byte
	resp.Body.Read(bodyBuffer)
	err = json.Unmarshal(bodyBuffer, &body)
	if err != nil {
		return nil, err
	}
	access.Catalog = ReadCatalog(body["token"])

	return &access, nil
}

func ReadCatalog(token tokenObject) Catalog {
	catalog := Catalog{}
	c, contains := token["catalog"]
	if !contains {
		return catalog
	}
	servicesList := c.(([]map[string]interface{}))
	for _, serviceDef := range servicesList {
		name := serviceDef["name"].(string)
		endpoints := serviceDef["endpoints"].([]endPointObject)
		for _, endpoint := range endpoints {
			if endpoint["interface"] == "public" {
				catalog[name] = Service{
					Id:     endpoint["id"],
					Name:   name,
					Url:    endpoint["url"],
					Region: endpoint["region"],
				}
			}
		}
	}
	return catalog
}
