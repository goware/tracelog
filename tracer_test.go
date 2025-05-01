package tracer

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTracer(t *testing.T) {
	tcr := NewTracerWithSizes(2, 2, 4)
	// tracer.Enable() // always enabled by default
	rawTcr := tcr.(*tracer)

	t.Run("trial 1", func(t *testing.T) {
		group := tcr.Group("server")
		trace := group.Span("run")
		trace.Info("boot")
		trace.Info("start")
		trace.Info("ready")

		time.Sleep(1000 * time.Millisecond)

		assertTrue(t, len(rawTcr.groupTS) == 1)
		assertTrue(t, len(rawTcr.spanTS["server"]) == 1)
		assertTrue(t, rawTcr.spanTS["server"]["run"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs) == 1)
		assertTrue(t, len(rawTcr.logs["server"]) == 1)
		assertTrue(t, len(rawTcr.logs["server"]["run"]) == 3)
	})

	time.Sleep(1500 * time.Millisecond)

	t.Run("trial 2", func(t *testing.T) {
		trace := tcr.Trace("api", "rpc")
		trace.Info("getUser")
		trace.Info("getProduct")
		trace.Info("getArticle")
		trace.Info("getFriend")
		trace.Info("getCity")

		time.Sleep(1000 * time.Millisecond)

		assertTrue(t, len(rawTcr.groupTS) == 2)
		assertTrue(t, len(rawTcr.spanTS) == 2)
		assertTrue(t, len(rawTcr.logs) == 2)

		assertTrue(t, len(rawTcr.spanTS["server"]) == 1)
		assertTrue(t, rawTcr.spanTS["server"]["run"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["server"]) == 1)
		assertTrue(t, len(rawTcr.logs["server"]["run"]) == 3)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 1)
		assertTrue(t, rawTcr.spanTS["api"]["rpc"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 1)
		assertTrue(t, len(rawTcr.logs["api"]["rpc"]) == 4)
	})

	time.Sleep(1000 * time.Millisecond)

	t.Run("trial 3 -- message overload", func(t *testing.T) {
		// we log more messages than the max allowed,
		// so some messages will be dropped
		trace := tcr.Trace("api", "db")
		trace.Info("getX")
		trace.Info("getY")
		trace.Info("setX")
		trace.Info("setY")
		time.Sleep(500 * time.Millisecond)
		trace.Warn("oops")
		time.Sleep(500 * time.Millisecond)
		trace.Error("boom")

		time.Sleep(1000 * time.Millisecond)

		assertTrue(t, len(rawTcr.groupTS) == 2)
		assertTrue(t, len(rawTcr.spanTS) == 2)
		assertTrue(t, len(rawTcr.logs) == 2)

		assertTrue(t, len(rawTcr.spanTS["server"]) == 1)
		assertTrue(t, rawTcr.spanTS["server"]["run"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["server"]) == 1)
		assertTrue(t, len(rawTcr.logs["server"]["run"]) == 3)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 2)
		assertTrue(t, rawTcr.spanTS["api"]["rpc"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 2)
		assertTrue(t, len(rawTcr.logs["api"]["rpc"]) == 4)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 2)
		assertTrue(t, rawTcr.spanTS["api"]["db"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 2)
		assertTrue(t, len(rawTcr.logs["api"]["db"]) == 4)

		// TODO: lets check the logs .. messages..
	})

	t.Run("trial 4 -- span overload", func(t *testing.T) {
		// we log more spans than the max allowed,
		// so some spans will be dropped. the new "cache" span
		trace := tcr.Trace("api", "cache")
		trace.Info("hitA")
		trace.Info("hitA")
		trace.Info("hitB")
		trace.Info("missA")
		trace.Info("missB")

		assertTrue(t, len(rawTcr.groupTS) == 2)
		assertTrue(t, len(rawTcr.spanTS) == 2)
		assertTrue(t, len(rawTcr.logs) == 2)

		assertTrue(t, len(rawTcr.spanTS["server"]) == 1)
		assertTrue(t, rawTcr.spanTS["server"]["run"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["server"]) == 1)
		assertTrue(t, len(rawTcr.logs["server"]["run"]) == 3)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 2)
		assertTrue(t, rawTcr.spanTS["api"]["db"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 2)
		assertTrue(t, len(rawTcr.logs["api"]["db"]) == 4)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 2)
		assertTrue(t, rawTcr.spanTS["api"]["cache"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 2)
		assertTrue(t, len(rawTcr.logs["api"]["cache"]) == 4)
	})

	time.Sleep(1000 * time.Millisecond)

	t.Run("trial 5 -- group overload", func(t *testing.T) {
		// we log more groups than the max allowed,
		// so some groups will be dropped. the new "jobqueue" group
		trace := tcr.Trace("jobqueue", "healthcheck")
		trace.Info("start")
		trace.Info("check 1")
		trace.Info("check 2")
		trace.Info("check 3")
		trace.Info("done")

		time.Sleep(1000 * time.Millisecond)

		assertTrue(t, len(rawTcr.groupTS) == 2)
		assertTrue(t, len(rawTcr.spanTS) == 2)
		assertTrue(t, len(rawTcr.logs) == 2)

		assertTrue(t, len(rawTcr.spanTS["jobqueue"]) == 1)
		assertTrue(t, rawTcr.spanTS["jobqueue"]["healthcheck"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["jobqueue"]) == 1)
		assertTrue(t, len(rawTcr.logs["jobqueue"]["healthcheck"]) == 4)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 2)
		assertTrue(t, rawTcr.spanTS["api"]["db"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 2)
		assertTrue(t, len(rawTcr.logs["api"]["db"]) == 4)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 2)
		assertTrue(t, rawTcr.spanTS["api"]["cache"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 2)
		assertTrue(t, len(rawTcr.logs["api"]["cache"]) == 4)
	})

	t.Run("trial 6", func(t *testing.T) {
		trace := tcr.Trace("api", "status")
		trace.Info("ok")
		trace.Warn("no")
		trace.Error("bad")
		trace.Info("ok")
		trace.Info("done")

		time.Sleep(1000 * time.Millisecond)

		assertTrue(t, len(rawTcr.groupTS) == 2)
		assertTrue(t, len(rawTcr.spanTS) == 2)
		assertTrue(t, len(rawTcr.logs) == 2)

		assertTrue(t, len(rawTcr.spanTS["jobqueue"]) == 1)
		assertTrue(t, rawTcr.spanTS["jobqueue"]["healthcheck"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["jobqueue"]) == 1)
		assertTrue(t, len(rawTcr.logs["jobqueue"]["healthcheck"]) == 4)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 2)
		assertTrue(t, rawTcr.spanTS["api"]["cache"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 2)
		assertTrue(t, len(rawTcr.logs["api"]["cache"]) == 4)

		assertTrue(t, len(rawTcr.spanTS["api"]) == 2)
		assertTrue(t, rawTcr.spanTS["api"]["status"].Before(time.Now()))
		assertTrue(t, len(rawTcr.logs["api"]) == 2)
		assertTrue(t, len(rawTcr.logs["api"]["status"]) == 4)
	})

	_, jsonOut := tcr.ToMap("EST", false, "", "")
	fmt.Println(string(jsonOut))
}

func TestTracerConcurrency(t *testing.T) {
	numGroups := 2
	numSpans := 2
	numMessages := 4
	tcr := NewTracerWithSizes(numGroups, numSpans, numMessages)
	rawTcr := tcr.(*tracer)

	numGoroutines := 20
	numGroupsToLog := 10   // More than numGroups limit
	numSpansToLog := 10    // More than numSpans limit
	numMessagesToLog := 20 // More than numMessages limit

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer wg.Done()
			for g := 0; g < numGroupsToLog; g++ {
				groupName := fmt.Sprintf("group-%d", g)
				for s := 0; s < numSpansToLog; s++ {
					spanName := fmt.Sprintf("span-%d", s)
					trace := tcr.Trace(groupName, spanName)
					for m := 0; m < numMessagesToLog; m++ {
						trace.Info("message from routine %d, group %d, span %d, msg %d", routineID, g, s, m)
						// Optional small sleep to encourage more interleaving,
						// though the lock serializes the critical section anyway.
						// time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	// Check limits
	if len(rawTcr.groupTS) != numGroups {
		t.Errorf("Expected %d groups, but got %d", numGroups, len(rawTcr.groupTS))
	}
	if len(rawTcr.logs) != numGroups {
		t.Errorf("Expected %d groups in logs map, but got %d", numGroups, len(rawTcr.logs))
	}

	// Check span and message limits for the remaining groups
	for groupName, spans := range rawTcr.logs {
		if _, ok := rawTcr.groupTS[groupName]; !ok {
			t.Errorf("Group '%s' exists in logs but not in groupTS", groupName)
		}
		if spanTimestamps, ok := rawTcr.spanTS[groupName]; ok {
			if len(spans) != numSpans {
				t.Errorf("Group '%s': Expected %d spans, but got %d", groupName, numSpans, len(spans))
			}
			if len(spanTimestamps) != numSpans {
				t.Errorf("Group '%s': Expected %d spans in spanTS, but got %d", groupName, numSpans, len(spanTimestamps))
			}

			// Check message limits for remaining spans
			for spanName, messages := range spans {
				if _, ok := spanTimestamps[spanName]; !ok {
					t.Errorf("Group '%s', Span '%s': exists in logs but not in spanTS", groupName, spanName)
				}
				if len(messages) > numMessages {
					t.Errorf("Group '%s', Span '%s': Expected max %d messages, but got %d", groupName, spanName, numMessages, len(messages))
				}
				if len(messages) == 0 && numMessages > 0 {
					// This shouldn't happen if we logged messages unless numMessages was 0
					t.Errorf("Group '%s', Span '%s': Found 0 messages, expected between 1 and %d", groupName, spanName, numMessages)
				}
			}

		} else {
			t.Errorf("Group '%s' exists in logs but not in spanTS map", groupName)
		}
	}
}
