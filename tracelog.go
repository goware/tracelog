package tracelog

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	DefaultGroupCount   = 40 // total groups
	DefaultSpanCount    = 60 // total spans per group
	DefaultMessageCount = 60 // total messages per span
)

type Tracelog interface {
	Trace(group, span string) Logger
	Group(group string) Logger

	ListGroups() []string
	ListSpans(group string) []string

	Logs(group string) [][]LogEntry
	ToMap(timezone string, withExactTime bool, groupFilter, spanFilter string) map[string]map[string][]string

	Enable()  // by default tracelog is enabled
	Disable() // disable all logging, turning each call into a noop
	IsEnabled() bool
}

type Logger interface {
	Span(span string) Logger
	With(group, span string) Logger

	GetGroup() string
	GetSpan() string

	Info(message string, v ...any)
	Warn(message string, v ...any)
	Error(message string, v ...any)
}

type LogEntry interface {
	Level() string
	Group() string
	Span() string
	Message() string
	Time() time.Time
	TimeAgo(timezone ...string) string
	Count() uint32
	FormattedMessage(timezone string, withExactTime ...bool) string
}

type tracelog struct {
	logs                             map[string]map[string][]logEntry
	numGroups, numSpans, numMessages int
	enabled                          bool
	groupTS                          map[string]time.Time
	spanTS                           map[string]time.Time
	mu                               sync.RWMutex
}

func NewTracelog() Tracelog {
	return NewTracelogWithSizes(DefaultGroupCount, DefaultSpanCount, DefaultMessageCount)
}

func NewTracelogWithSizes(numGroups, numSpans, numMessages int) Tracelog {
	if numGroups < 1 {
		numGroups = DefaultGroupCount
	}
	if numSpans < 1 {
		numSpans = DefaultSpanCount
	}
	if numMessages < 1 {
		numMessages = DefaultMessageCount
	}

	return &tracelog{
		logs:        make(map[string]map[string][]logEntry),
		numGroups:   numGroups,
		numSpans:    numSpans,
		numMessages: numMessages,
		enabled:     true,
		groupTS:     make(map[string]time.Time),
		spanTS:      make(map[string]time.Time),
	}
}

func (t *tracelog) Trace(group, span string) Logger {
	return &logger{
		tracelog: t,
		group:    group,
		span:     span,
	}
}

func (t *tracelog) Group(group string) Logger {
	return &logger{
		tracelog: t,
		group:    group,
	}
}

func (t *tracelog) ListGroups() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	groups := make([]string, 0, len(t.logs))
	for group := range t.logs {
		groups = append(groups, group)
	}
	return groups
}

func (t *tracelog) ListSpans(group string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	spans := make([]string, 0, len(t.logs[group]))
	for span := range t.logs[group] {
		spans = append(spans, span)
	}
	return spans
}

func (t *tracelog) Logs(group string) [][]LogEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if _, ok := t.logs[group]; !ok {
		return [][]LogEntry{}
	}

	spans := make([]string, 0, len(t.logs[group]))
	for span := range t.logs[group] {
		spans = append(spans, span)
	}

	sort.Slice(spans, func(i, j int) bool {
		timeI := t.spanTS[spans[i]]
		timeJ := t.spanTS[spans[j]]
		return timeI.After(timeJ) // most recent first
	})

	out := make([][]LogEntry, 0, len(spans))
	for _, span := range spans {
		entries := t.logs[group][span]
		outSpan := make([]LogEntry, 0, len(entries))
		for _, entry := range entries {
			outSpan = append(outSpan, entry)
		}
		out = append(out, outSpan)
	}

	return out
}

func (t *tracelog) ToMap(timezone string, withExactTime bool, groupFilter, spanFilter string) map[string]map[string][]string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	m := make(map[string]map[string][]string)

	for group, spans := range t.logs {
		if groupFilter != "" && !strings.HasPrefix(group, groupFilter) {
			continue
		}

		groupMap := make(map[string][]string)
		for span, entries := range spans {
			if spanFilter != "" && !strings.HasPrefix(span, spanFilter) {
				continue
			}

			for _, entry := range entries {
				groupMap[span] = append(groupMap[span], entry.FormattedMessage(timezone, withExactTime))
			}
		}

		m[group] = groupMap
	}

	return m
}

func (t *tracelog) Enable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = true
}

func (t *tracelog) Disable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = false
}

func (t *tracelog) IsEnabled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.enabled
}

type logger struct {
	tracelog *tracelog
	group    string
	span     string
}

var _ Logger = &logger{}

func (l *logger) Span(span string) Logger {
	return &logger{
		tracelog: l.tracelog,
		group:    l.group,
		span:     span,
	}
}

func (l *logger) With(group, span string) Logger {
	return &logger{
		tracelog: l.tracelog,
		group:    group,
		span:     span,
	}
}

func (l *logger) GetGroup() string {
	return l.group
}

func (l *logger) GetSpan() string {
	return l.span
}

func (l *logger) Info(message string, v ...any) {
	l.log("INFO", l.group, l.span, message, v...)
}

func (l *logger) Warn(message string, v ...any) {
	l.log("WARN", l.group, l.span, message, v...)
}

func (l *logger) Error(message string, v ...any) {
	l.log("ERROR", l.group, l.span, message, v...)
}

func (l *logger) log(level, group, span, message string, v ...any) {
	if !l.tracelog.IsEnabled() {
		return
	}

	l.tracelog.mu.Lock()
	defer l.tracelog.mu.Unlock()

	// TODO: limit the number of groups ..
	// which one do we bump..? we need a timestamp per group..
	// and timestamp per span ..

	// update timestamps
	l.tracelog.groupTS[group] = time.Now().UTC()
	l.tracelog.spanTS[span] = time.Now().UTC()

	// add log entry
	g := l.tracelog.logs[group]
	if g == nil {
		g = make(map[string][]logEntry)
	}

	s := g[span]
	if s == nil {
		s = make([]logEntry, 0, l.tracelog.numMessages)
	}

	msg := fmt.Sprintf(message, v...)
	if len(msg) == 0 {
		return
	}
	if len(msg) > 1000 {
		// truncate message to 1000 characters
		msg = msg[:1000]
	}

	// First check if the message is already in the log, and if so, increment the count
	// and update the time
	for i, entry := range s {
		if entry.message == msg && entry.level == level {
			entry.count++
			entry.time = time.Now().UTC()
			s[i] = entry
			l.tracelog.logs[group][span] = s
			sort.Slice(s, func(i, j int) bool {
				return s[i].time.After(s[j].time)
			})
			return
		}
	}

	// Add the new entry
	newLen := len(s) + 1
	if newLen > l.tracelog.numMessages {
		newLen = l.tracelog.numMessages
	}
	newS := make([]logEntry, newLen)
	newEntry := logEntry{
		group:   l.group,
		span:    l.span,
		message: msg,
		level:   level,
		time:    time.Now().UTC(),
		count:   1,
	}
	newS[0] = newEntry
	if len(s) > 0 {
		copy(newS[1:], s[:newLen-1])
	}

	l.tracelog.logs[group] = g
	l.tracelog.logs[group][span] = newS
}

type logEntry struct {
	group   string
	span    string
	message string
	level   string
	time    time.Time
	count   uint32
}

var _ LogEntry = logEntry{}

func (l logEntry) Group() string {
	return l.group
}

func (l logEntry) Span() string {
	return l.span
}

func (l logEntry) Message() string {
	return l.message
}

func (l logEntry) Level() string {
	return l.level
}

func (l logEntry) Time() time.Time {
	return l.time
}

func (l logEntry) TimeAgo(timezone ...string) string {
	var err error
	loc := time.UTC
	if len(timezone) > 0 {
		loc, err = time.LoadLocation(timezone[0])
		if err != nil {
			loc = time.UTC
		}
	}

	duration := time.Since(l.time.In(loc))

	if duration < time.Minute {
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		seconds := int(duration.Seconds()) % 60
		return fmt.Sprintf("%dm %ds ago", int(duration.Minutes()), seconds)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		if minutes == 0 {
			return fmt.Sprintf("%dh ago", hours)
		}
		return fmt.Sprintf("%dh %dm ago", hours, minutes)
	}

	return l.time.In(loc).Format(time.RFC822)
}

func (l logEntry) Count() uint32 {
	return l.count
}

func (l logEntry) FormattedMessage(timezone string, withExactTime ...bool) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	var out string
	if len(withExactTime) > 0 && withExactTime[0] {
		out = fmt.Sprintf("%s - [%s] %s", l.time.In(loc).Format(time.RFC822), l.level, l.message)
	} else {
		out = fmt.Sprintf("%s - [%s] %s", l.TimeAgo(timezone), l.level, l.message)
	}
	if l.count > 1 {
		return fmt.Sprintf("%s [x%d]", out, l.count)
	} else {
		return out
	}
}
