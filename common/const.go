package common

import (
	"regexp"
)

const (
	PodStateUnknown       = 0
	PodWaitingForSchedule = 1
	PodScheduled          = 2
	PodReady              = 3
	PodTerminated         = 5
)

var ReMail *regexp.Regexp

const (
	SAEStaticPath     = "api/sae/static"
	SAERuntimePath    = SAEStaticPath + "/runtime"
	SCMDynamicJobPath = "api/scm/builds/jobs.yml"
)
