package main

import (
	rice "github.com/GeertJohan/go.rice"
	"github.com/pflow-dev/pflow/cmd"
	"github.com/pflow-dev/pflow/service"
)

func init() {
	service.Box = rice.MustFindBox("./public")
}

func main() {
	cmd.Execute()
}
