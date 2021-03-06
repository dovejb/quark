package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dovejb/quark"
)

var (
	q = quark.NewQuark()
)

type example struct {
	quark.Console
}

func (s example) Vehicle_Vin() (rsp string) {
	return "Vehicle_Vin"
}

func (s example) Vehicle_vins(vin string, params struct {
	Name string
}) (rsp string) {
	return fmt.Sprintf("Vehicle_vin as string, %s", vin)
}

/*
func (s root) Vehicle_vini(vin int) (rsp string) {
	return fmt.Sprintf("Vehicle_vin as int, %d", vin)
}

func (s root) Vehicle_vinf(vin float64) (rsp string) {
	return fmt.Sprintf("Vehicle_vin as float64, %f", vin)
}
*/

type User struct {
	Name  string
	Title string
	Index int
}

func (s example) Users() (rsp []User) {
	return
}
func (s example) GET_Vehicle_groupId_vin(groupId int, vin string) (rsp struct {
	Vin   string
	Name  string
	Admin string
}) {
	log.Println("GET_Vehicle_groupId_vin")
	return
}

func (s example) PATCH_Vehicle_groupId_vin(groupId int, vin string, req struct {
	Name  string
	Admin string
}) {
	log.Println("PATCH_Vehicle_groupId_vin")
	return
}

func (s example) Hello_World() string {
	return "hello world!"
}

func (s example) JsonResponse() (rsp struct {
	Name      string
	Corp      string
	Education []struct {
		Type       string
		SchoolName string
		From       string
		To         string
	}
}) {
	rsp.Name = "dovejb"
	rsp.Corp = "abc"
	quark.MakeSlice(&rsp.Education, 1)
	rsp.Education[0].Type = "elementary"
	rsp.Education[0].SchoolName = "panda school"
	rsp.Education[0].From = "2018"
	rsp.Education[0].To = "2021"
	quark.Resize(&rsp.Education, 2)
	return
}

func (s example) InternalServerError() {
	s.Halt(http.StatusInternalServerError, "internal_server_error")
}

func (s example) Panic() {
	panic("baga")
}

func (s example) Full_Parameters_pathv_type_TryIt(pathv string, typ int, req struct {
	Value   quark.Int
	Price   quark.Number
	Message quark.String
	Hello   string
	World   struct {
		Oceans     []string
		Continents []string
		Air        string
	}
}) (rsp []User) {
	fmt.Println(s.Body())
	fmt.Println(quark.Js(req))
	return
}

func (s example) NoRequestButWithBody() []byte {
	fmt.Println(s.Body())
	return s.Body()
}

func Authenticate(c *quark.Console) bool {
	r := c.Request()
	if r.Header.Get("Authorization") == "dovejb" {
		return true
	}
	return false
}

func main() {
	q.RegisterService(example{})
	//q.WithAuthenticate(Authenticate)
	q.WithPathPrefix([]string{"open", "v1"})
	fmt.Println(quark.Js(q.SwaggerSpec()))
	for i := range q.Services {
		log.Println(q.Services[i])
		q.Services[i].DumpPaths()
	}
	http.ListenAndServe(":11019", q)
}
