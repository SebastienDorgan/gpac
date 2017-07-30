package openstack

import "fmt"

type AuthAPIV2 struct {
	URL          string
	UserId       string
	UserPassword string
	ProjectId    string
}

func (auth AuthAPIV2) AuthRequest() (string) {
	request := `
	{
	  "auth":{
		"tenantName":"%q",
		"passwordCredentials":{
		  "username":"%d",
		  "password":"%q"
         }
      }
    }
    `
	fmt.Sprintf(request,
		auth.ProjectId, auth.UserId, auth.UserPassword, )
	return request
}

func (AuthAPIV2) Authenticate() (AccessData, error) {
	panic("implement me")
}
