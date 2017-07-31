package openstack

import (
	"encoding/json"
	"fmt"
	"github.com/SebastienDorgan/gpac/cloud/openstack"
	"testing"
)

func TestAuthRequestPasswordUnScoped(t *testing.T) {
	userName := "test"
	userPassword := "test_password"
	auth := openstack.AuthAPIV3{
		UserName:     &userName,
		UserPassword: &userPassword,
	}
	request := auth.AuthRequest()

	type User struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	type Password struct {
		User User `json:"user"`
	}
	type Identity struct {
		Methods  []string `json:"methods"`
		Password Password `json:"password"`
	}
	type Auth struct {
		Identity Identity `json:"identity"`
	}
	type ObjTest struct {
		Auth Auth `json:"auth"`
	}

	obj := ObjTest{}
	err := json.Unmarshal([]byte(request), &obj)
	if err != nil {
		t.Errorf(fmt.Sprint(err))
		t.Fail()
	}
	if len(obj.Auth.Identity.Methods) != 1 {
		t.Errorf("Auth.Identity.Methods length should be one")
		t.Fail()
	}
	if obj.Auth.Identity.Methods[0] != "password" {
		t.Errorf("Auth.Identity.Methods[0] should be equal to 'password'")
		t.Fail()
	}
	if obj.Auth.Identity.Password.User.Name != "test" {
		t.Errorf("Auth.Identity.Password.User.Name should be 'test'")
		t.Fail()
	}
	if obj.Auth.Identity.Password.User.Password != "test_password" {
		t.Errorf("Auth.Identity.Password.User.Password should be 'test_password'")
		t.Fail()
	}
}
func TestAuthRequestPasswordUnScopedWithUserDomainName(t *testing.T) {
	userName := "test"
	userPassword := "test_password"
	userDomainName := "my.domain"
	auth := openstack.AuthAPIV3{
		UserName:       &userName,
		UserPassword:   &userPassword,
		UserDomainName: &userDomainName,
	}
	request := auth.AuthRequest()

	type UserDomain struct {
		Name string `json:"name"`
	}

	type User struct {
		Name     string     `json:"name"`
		Password string     `json:"password"`
		Domain   UserDomain `json:"domain"`
	}
	type Password struct {
		User User `json:"user"`
	}
	type Identity struct {
		Methods  []string `json:"methods"`
		Password Password `json:"password"`
	}
	type Auth struct {
		Identity Identity `json:"identity"`
	}
	type ObjTest struct {
		Auth Auth `json:"auth"`
	}

	obj := ObjTest{}
	err := json.Unmarshal([]byte(request), &obj)
	if err != nil {
		t.Errorf(fmt.Sprint(err))
		t.Fail()
	}
	if len(obj.Auth.Identity.Methods) != 1 {
		t.Errorf("Auth.Identity.Methods length should be one")
		t.Fail()
	}
	if obj.Auth.Identity.Methods[0] != "password" {
		t.Errorf("obj.Auth.Identity.Methods[0] should be equal to 'password'")
		t.Fail()
	}
	if obj.Auth.Identity.Password.User.Name != "test" {
		t.Errorf("Auth.Identity.Password.User.Name should be 'test'")
		t.Fail()
	}
	if obj.Auth.Identity.Password.User.Password != "test_password" {
		t.Errorf("Auth.Identity.Password.User.Password should be 'test_password'")
		t.Fail()
	}
	if obj.Auth.Identity.Password.User.Domain.Name != "my.domain" {
		t.Errorf("Auth.Identity.Password.User.Domain.Name should be 'my.domain'")
		t.Fail()
	}

}

func TestAuthRequestPasswordUnScopedWithUserDomainId(t *testing.T) {
	userName := "test"
	userPassword := "test_password"
	userDomainName := "my.domain"
	auth := openstack.AuthAPIV3{
		UserName:     &userName,
		UserPassword: &userPassword,
		UserDomainId: &userDomainName,
	}
	request := auth.AuthRequest()

	type UserDomain struct {
		Id string `json:"id"`
	}

	type User struct {
		Name     string     `json:"name"`
		Password string     `json:"password"`
		Domain   UserDomain `json:"domain"`
	}
	type Password struct {
		User User `json:"user"`
	}
	type Identity struct {
		Methods  []string `json:"methods"`
		Password Password `json:"password"`
	}
	type Auth struct {
		Identity Identity `json:"identity"`
	}
	type ObjTest struct {
		Auth Auth `json:"auth"`
	}

	obj := ObjTest{}
	err := json.Unmarshal([]byte(request), &obj)
	if err != nil {
		t.Errorf(fmt.Sprint(err))
		t.Fail()
	}
	if len(obj.Auth.Identity.Methods) != 1 {
		t.Errorf("Auth.Identity.Methods length should be one")
		t.Fail()
	}
	if obj.Auth.Identity.Methods[0] != "password" {
		t.Errorf("obj.Auth.Identity.Methods[0] should be equal to 'password'")
		t.Fail()
	}
	if obj.Auth.Identity.Password.User.Name != "test" {
		t.Errorf("Auth.Identity.Password.User.Name should be 'test'")
		t.Fail()
	}
	if obj.Auth.Identity.Password.User.Password != "test_password" {
		t.Errorf("Auth.Identity.Password.User.Password should be 'test_password'")
		t.Fail()
	}
	if obj.Auth.Identity.Password.User.Domain.Id != "my.domain" {
		t.Errorf("Auth.Identity.Password.User.Domain.Id should be 'my.domain'")
		t.Fail()
	}

}

func TestAuthRequestExplicitlyUnScoped(t *testing.T) {
	userName := "test"
	userPassword := "test_password"
	auth := openstack.AuthAPIV3{
		UserName:          &userName,
		UserPassword:      &userPassword,
		ExplicitlyUnScope: true,
	}
	request := auth.AuthRequest()

	type User struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	type Password struct {
		User User `json:"user"`
	}
	type Identity struct {
		Methods  []string `json:"methods"`
		Password Password `json:"password"`
	}

	type Auth struct {
		Identity Identity `json:"identity"`
		Scope    string   `json:"scope"`
	}
	type ObjTest struct {
		Auth Auth `json:"auth"`
	}

	obj := ObjTest{}
	err := json.Unmarshal([]byte(request), &obj)
	if err != nil {
		t.Errorf(fmt.Sprint(err))
		t.Fail()
	}

	if obj.Auth.Scope != "unscoped" {
		t.Errorf("Auth.Scope should be equal to 'unscoped'")
		t.Fail()
	}

}

func TestAuthRequestScopedWithId(t *testing.T) {
	userName := "test"
	userPassword := "test_password"
	projectId := "test-project.id"
	auth := openstack.AuthAPIV3{
		UserName:     &userName,
		UserPassword: &userPassword,
		ProjectId:    &projectId,
	}
	request := auth.AuthRequest()
	fmt.Println(request)

	type User struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	type Password struct {
		User User `json:"user"`
	}
	type Identity struct {
		Methods  []string `json:"methods"`
		Password Password `json:"password"`
	}
	type Project struct {
		Id string `json:"id"`
	}
	type Scope struct {
		Project Project `json:"project"`
	}
	type Auth struct {
		Identity Identity `json:"identity"`
		Scope    Scope    `json:"scope"`
	}
	type ObjTest struct {
		Auth Auth `json:"auth"`
	}

	obj := ObjTest{}
	err := json.Unmarshal([]byte(request), &obj)
	if err != nil {
		t.Errorf(fmt.Sprint(err))
		t.Fail()
	}

	if obj.Auth.Scope.Project.Id != "test-project.id" {
		t.Errorf("Auth.Scope.Project.Id should be equal to 'test-project.id'")
		t.Fail()
	}

}

func TestAuthRequestScopedWithNameAndDomainName(t *testing.T) {
	userName := "test"
	userPassword := "test_password"
	projectName := "test-project.name"
	projectDomain := "test-project.domain"
	auth := openstack.AuthAPIV3{
		UserName:          &userName,
		UserPassword:      &userPassword,
		ProjectName:       &projectName,
		ProjectDomainName: &projectDomain,
	}
	request := auth.AuthRequest()
	fmt.Println(request)

	type User struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	type Password struct {
		User User `json:"user"`
	}
	type Identity struct {
		Methods  []string `json:"methods"`
		Password Password `json:"password"`
	}

	type Domain struct {
		Name string `json:"name"`
	}
	type Project struct {
		Name   string `json:"name"`
		Domain Domain `json:"domain"`
	}
	type Scope struct {
		Project Project `json:"project"`
	}
	type Auth struct {
		Identity Identity `json:"identity"`
		Scope    Scope    `json:"scope"`
	}
	type ObjTest struct {
		Auth Auth `json:"auth"`
	}

	obj := ObjTest{}
	err := json.Unmarshal([]byte(request), &obj)
	if err != nil {
		t.Errorf(fmt.Sprint(err))
		t.Fail()
	}

	if obj.Auth.Scope.Project.Name != "test-project.name" {
		t.Errorf("Auth.Scope.Project.Name should be equal to 'test-project.name'")
		t.Fail()
	}

	if obj.Auth.Scope.Project.Domain.Name != "test-project.domain" {
		t.Errorf("Auth.Scope.Project.Domain.Name should be equal to 'test-project.domain'")
		t.Fail()
	}

}

func TestAuthRequestScopedWithNameAndDomainId(t *testing.T) {
	userName := "test"
	userPassword := "test_password"
	projectName := "test-project.name"
	//projectDomain := "test-project.domain"
	auth := openstack.AuthAPIV3{
		UserName:          &userName,
		UserPassword:      &userPassword,
		ProjectName:       &projectName,
		ProjectDomainName: &projectName,
	}
	request := auth.AuthRequest()
	fmt.Println(request)

	type User struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	type Password struct {
		User User `json:"user"`
	}
	type Identity struct {
		Methods  []string `json:"methods"`
		Password Password `json:"password"`
	}
	type Project struct {
		Name string `json:"name"`
	}
	type Scope struct {
		Project Project `json:"project"`
	}
	type Auth struct {
		Identity Identity `json:"identity"`
		Scope    Scope    `json:"scope"`
	}
	type ObjTest struct {
		Auth Auth `json:"auth"`
	}

	obj := ObjTest{}
	err := json.Unmarshal([]byte(request), &obj)
	if err != nil {
		t.Errorf(fmt.Sprint(err))
		t.Fail()
	}

	if obj.Auth.Scope.Project.Name != "test-project.name" {
		t.Errorf("Auth.Scope.Project.Name should be equal to 'test-project.name'")
		t.Fail()
	}

}

func TestReadCatalog(t *testing.T) {
	//userName := "test"
	//userPassword := "test_password"
	////auth := openstack.AuthAPIV3{
	////	UserName:          &userName,
	////	UserPassword:      &userPassword,
	////}
	////openstack.ReadCatalog()

}
