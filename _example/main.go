package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/goware/tracelog"
)

func main() {
	tracelog := tracelog.NewTracelog()

	group := tracelog.Group("server")
	trace := group.Span("run")

	trace.Info("boot")
	trace.Info("start")
	trace.Info("ready")

	group2 := tracelog.Group("api")
	trace2 := group2.Span("run")
	trace2.Info("load")
	trace2.Info("fetch")

	spew.Dump(tracelog.ListGroups())
	spew.Dump(tracelog.Logs("server"))
	spew.Dump(tracelog.Logs("api"))
}

// func main() {
// 	ql := quicklog.NewQuicklog()

// 	ql.Info("g1", "hi")
// 	ql.Info("g1", "test")
// 	ql.Warn("g1", "test")
// 	ql.Warn("g1", "test")
// 	time.Sleep(3 * time.Second)
// 	for _, l := range ql.Logs("g1") {
// 		fmt.Println(l.FormattedMessage("EST", false))
// 	}

// 	ql.Info("g1", "test")
// 	ql.Warn("g1", "test2")
// 	time.Sleep(10 * time.Second)
// 	for _, l := range ql.Logs("g1") {
// 		fmt.Println(l.FormattedMessage("EST", false))
// 	}

// 	ql.Warn("g1", "test3")
// 	time.Sleep(100 * time.Second)
// 	for _, l := range ql.Logs("g1") {
// 		fmt.Println(l.FormattedMessage("EST", false))
// 	}

// }

// func main2() {
// 	ql := quicklog.NewQuicklog(40, 60, 60) // number of: groups, spans, messages

// 	tracer := tracelog.NewTracer(40, 60, 60)

// 	group := ql.Group("server")
// 	trace := group.Span("run")

// 	trace.Info("boot")
// 	trace.Info("start")
// 	trace.Info("ready")

// 	//--

// 	group := ql.Group("server")
// 	trace := group.Trace("run")

// 	span.Info("boot")
// 	span.Info("start")
// 	span.Info("ready")

// }
