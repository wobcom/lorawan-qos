package main

import (
	"network-qos/cmd/run"
)

var version string // set by the compiler

func main() {
	run.Execute(version)
}
