package main

import (
	"time"

	"github.com/goware/tracelog"
)

func main() {
	tracelog := tracelog.NewTracelogWithSizes(2, 2, 4)
	// tracelog.Enable() // always enabled by default
	// tracelog.Disable() // disable all logging, turning each call into a noop

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
	}

	time.Sleep(1500 * time.Millisecond)

	{
		trace := tracelog.Trace("api", "rpc")
		trace.Info("getUser")
		trace.Info("getProduct")
		trace.Info("getArticle")
		trace.Info("getFriend")
		trace.Info("getCity")

		time.Sleep(1000 * time.Millisecond)

		trace = tracelog.Trace("api", "db")
		trace.Info("getX")
		trace.Info("getY")
		trace.Info("setX")
		trace.Info("setY")
		trace.Warn("oops")
		trace.Error("boom")

		time.Sleep(1000 * time.Millisecond)

		trace = tracelog.Trace("api", "cache")
		trace.Info("hitA")
		trace.Info("hitA")
		trace.Info("hitB")
		trace.Info("missA")
		trace.Info("missB")
		trace.Info("hitA")
	}

	// {
	// 	time.Sleep(1000 * time.Millisecond)
	// 	trace := tracelog.Trace("probe", "status")
	// 	trace.Info("ok 1")
	// 	trace.Info("ok 2")
	// 	trace.Info("ok 3")
	// 	trace.Info("ok 4")
	// 	trace.Info("ok 5")
	// }

	time.Sleep(1000 * time.Millisecond)

	{
		trace := tracelog.Trace("jobqueue", "healthcheck")
		trace.Info("start")
		trace.Info("check 1")
		trace.Info("check 2")
		trace.Info("check 3")
		trace.Info("done")
	}

	time.Sleep(1000 * time.Millisecond)

	// spew.Dump(tracelog.ListGroups())
	// spew.Dump(tracelog.Logs("server"))
	// spew.Dump(tracelog.Logs("api"))
	// spew.Dump(tracelog.Logs("jobqueue"))
	// fmt.Println("\n----\n")
	// spew.Dump(tracelog.ToMap("EST", false, "", ""))
}
