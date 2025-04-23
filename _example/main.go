package main

import (
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/goware/tracelog"
)

func main() {
	tracelog := tracelog.NewTracelog()

	{
		group := tracelog.Group("server")
		trace := group.Span("run")

		trace.Info("boot")
		trace.Info("start")
		trace.Info("ready")

		time.Sleep(1000 * time.Millisecond)

		trace = group.Span("startServices")
		trace.Info("startAPI")
		trace.Info("startCache")

		time.Sleep(1500 * time.Millisecond)
	}

	{
		trace := tracelog.Trace("api", "rpc")
		trace.Info("getUser")
		trace.Info("getProduct")

		time.Sleep(1000 * time.Millisecond)
	}

	spew.Dump(tracelog.ListGroups())
	spew.Dump(tracelog.Logs("server"))
	spew.Dump(tracelog.Logs("api"))

	fmt.Println("\n----\n")

	spew.Dump(tracelog.ToMap("EST", false, "", ""))
}
