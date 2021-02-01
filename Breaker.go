package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"net/http"
)

var (
	breakerLog = logrus.WithFields(logrus.Fields{
		"module": "GoBreaker",
		"file":   "Breaker.go",
	})
)

func NewPathBreakers() PathBreakers {
	return make(PathBreakers)
}

type PathBreakers map[string]*gobreaker.CircuitBreaker

func (pb PathBreakers) GetBreakerSettingForRequest(req *http.Request) gobreaker.Settings {
	completePath := req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		completePath = fmt.Sprintf("%s?%s", completePath, req.URL.RawQuery)
	}
	return gobreaker.Settings{
		Name:        completePath,
		MaxRequests: 0,
		Interval:    0,
		Timeout:     0,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			breakerLog.Infof("[%s] changed state from %s to %s", from.String(), to.String())
		},
	}
}

func (pb PathBreakers) GetBreakerForRequest(req *http.Request) *gobreaker.CircuitBreaker {
	completePath := req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		completePath = fmt.Sprintf("%s?%s", completePath, req.URL.RawQuery)
	}
	if b, ok := pb[completePath]; ok {
		return b
	}
	newBreaker := gobreaker.NewCircuitBreaker(pb.GetBreakerSettingForRequest(req))
	pb[completePath] = newBreaker
	return newBreaker
}
