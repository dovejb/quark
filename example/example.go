package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dovejb/quark"
	"github.com/dovejb/quark/types"
)

var (
	q = quark.NewQuark()
)

type root struct {
	quark.Console
}

func (s root) Vehicle_Vin() (rsp string) {
	return "Vehicle_Vin"
}

func (s root) Vehicle_vins(vin string, params struct {
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

func (s root) Users() (rsp []User) {
	return
}
func (s root) GET_Vehicle_groupId_vin(groupId int, vin string) (rsp struct {
	Vin   string
	Name  string
	Admin string
}) {
	log.Println("GET_Vehicle_groupId_vin")
	return
}

func (s root) PATCH_Vehicle_groupId_vin(groupId int, vin string, req struct {
	Name  string
	Admin string
}) {
	log.Println("PATCH_Vehicle_groupId_vin")
	return
}

func (s root) Hello_World() string {
	return "hello world!"
}

func (s root) JsonResponse() (rsp struct {
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

func (s root) InternalServerError() {
	s.Halt(http.StatusInternalServerError, "internal_server_error")
}

func (s root) Panic() {
	panic("baga")
}

func (s root) Full_Parameters_pathv_type_TryIt(pathv string, typ int, req struct {
	Value   types.URLInt
	Price   types.URLNumber
	Message types.URLString
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

func (s root) NoRequestButWithBody() []byte {
	fmt.Println(s.Body())
	return s.Body()
}

func main() {
	q.RegisterService(root{})
	fmt.Println(quark.Js(q.SwaggerSpec()))
	for i := range q.Services {
		log.Println(q.Services[i])
		q.Services[i].DumpPaths()
	}
	http.ListenAndServe(":11019", q)
}
