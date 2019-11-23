package common

type FlowStage struct {
	Name   string      `json:"name"`
	Prompt string      `json:"prompt"`
	State  uint        `json:"status"`
	Extra  interface{} `json:"extra"`
}

const (
	FlowStageWait      = 0
	FlowStageInProcess = 1
	FlowStagePassed    = 2
	FlowStageRejected  = 3
	FlowStageError     = 4
	FlowStageSkip      = 5
)
