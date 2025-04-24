package main

import (
	"time"

	"github.com/goware/tracer"
)

func main() {
	tracer := tracer.NewTracerWithSizes(2, 2, 4)
	// tracer.Enable() // always enabled by default
	// tracer.Disable() // disable all logging, turning each call into a noop

	{
		group := tracer.Group("server")
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
		trace := tracer.Trace("api", "rpc")
		trace.Info("getUser")
		trace.Info("getProduct")
		trace.Info("getArticle")
		trace.Info("getFriend")
		trace.Info("getCity")

		time.Sleep(1000 * time.Millisecond)

		trace = tracer.Trace("api", "db")
		trace.Info("getX")
		trace.Info("getY")
		trace.Info("setX")
		trace.Info("setY")
		time.Sleep(500 * time.Millisecond)
		trace.Warn("oops")
		time.Sleep(500 * time.Millisecond)
		trace.Error("boom")

		time.Sleep(1000 * time.Millisecond)

		trace = tracer.Trace("api", "cache")
		trace.Info("hitA")
		trace.Info("hitA")
		trace.Info("hitB")
		trace.Info("missA")
		trace.Info("missB")
		trace.Info("hitA")
	}

	// {
	// 	time.Sleep(1000 * time.Millisecond)
	// 	trace := tracer.Trace("probe", "status")
	// 	trace.Info("ok 1")
	// 	trace.Info("ok 2")
	// 	trace.Info("ok 3")
	// 	trace.Info("ok 4")
	// 	trace.Info("ok 5")
	// }

	time.Sleep(1000 * time.Millisecond)

	{
		trace := tracer.Trace("jobqueue", "healthcheck")
		trace.Info("start")
		trace.Info("check 1")
		trace.Info("check 2")
		trace.Info("check 3")
		trace.Info("done")
	}

	time.Sleep(1000 * time.Millisecond)

	// spew.Dump(tracer.ListGroups())
	// spew.Dump(tracer.Logs("server"))
	// spew.Dump(tracer.Logs("api"))
	// spew.Dump(tracer.Logs("jobqueue"))
	// fmt.Println("\n----\n")
	// spew.Dump(tracer.ToMap("EST", false, "", ""))
}
