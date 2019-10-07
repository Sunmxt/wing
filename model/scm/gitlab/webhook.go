package gitlab

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	MergeRequestClosed  = 1
	MergeRequestOpen   = 2
	MergeRequestMerged = 3
	MergeRequestReopen = 4
)

type MergeRequestEvent struct {
	Event int

	User
	SourceProject *Project
	TargetProject *Project
	MergeRequest
}

type RawMergeRequestEventProject struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	WebURL            string `json:"web_url"`
	AvatarURL         string `json:"avatar_url"`
	GitSSHURL         string `json:"git_ssh_url"`
	GitHTTPURL        string `json:"git_http_url"`
	Namespace         string `json:"namespace"`
	VisibilityLevel   uint   `json:"visibility_level"`
	PathWithNamespace string `json:"path_with_namespace"`
	DefaultBranch     string `json:"default_branch"`
	CIConfigPath      string `json:"ci_config_path"`
	Homepage          string `json:"homepage"`
	URL               string `json:"url"`
	SSHURL            string `json:"ssh_url"`
	HTTPURL           string `json:"http_url"`
}

type RawMergeRequestEvent struct {
	ObjectKind string `json:"object_kind"`
	User       `json:"user"`
	Project    *RawMergeRequestEventProject `json:"project"`

	ObjectAttrs struct {
		ID              uint   `json:"id"`
		TargetBranch    string `json:"target_branch"`
		SourceBranch    string `json:"source_branch"`
		SourceProjectID uint   `json:"source_project_id"`
		AuthorID        uint   `json:"author_id"`
		AssigneeID      uint   `json:"assignee_id"`
		Title           string `json:"title"`
		CreateAt        string `json:"create_at"`
		UpdateAt        string `json:"update_at"`
		MilestoneID     uint   `json:"milestone_id"`
		State           string `json:"state"`
		MergeStatus     string `json:"merge_status"`
		TargetProjectID uint   `json:"target_project_id"`
		InternalID      uint   `json:"iid"`
		Description     string `json:"description"`
		UpdatedByID     uint   `json:"updated_by_id"`
		//MergeError
		MergeParams struct {
			ForceRemoveSourceBranch string `json:"force_remove_source_branch"`
		} `json:"merge_params"`
		MergeWhenPipelineSucceeds bool                         `json:"merge_when_pipeline_succeeds"`
		MergeUserID               uint                         `json:"merge_user_id"`
		MergeCommitSHA            string                       `json:"merge_commit_sha"`
		DeleteAt                  string                       `json:"delete_at"`
		InProgressMergeCommitSHA  string                       `json:"in_progress_merge_commit_sha"`
		LockVersion               uint                         `json:"lock_version"`
		TimeEstimate              uint                         `json:"time_estimate"`
		LastEditedAt              string                       `json:"LastEditedAt"`
		LastEditedByID            uint                         `json:"last_edited_by_id"`
		HeadPipelineID            uint                         `json:"head_pipeline_id"`
		RefFetched                bool                         `json:"ref_fetched"`
		MergeJID                  uint                         `json:"merge_jid"`
		Source                    *RawMergeRequestEventProject `json:"source"`
		Target                    *RawMergeRequestEventProject `json:"target"`
		LastCommit                struct {
			ID        string `json:"id"`
			Message   string `json:"message"`
			Timestamp string `json:"timestamp"`
			URL       string `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"last_commit"`
		WorkInProgress      bool   `json:"work_in_progress"`
		TotalTimeSpent      uint   `json:"total_time_spent"`
		HumanTotalTimeSpent uint   `json:"human_total_time_spent"`
		HumanTimeEstimate   uint   `json:"human_total_time_estimate"`
		Action              string `json:"action"`
	} `json:"object_attributes"`
	//Labels     []string `json:"labels"`
	Repository struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Homepage    string `json:"homepage"`
	} `json:"repository"`
}

type EventObjectKind struct {
	ObjectKind string `json:"object_kind"`
}

type MergeRequestEventHandler func(*http.Request, *MergeRequestEvent) error

var typeMergeRequestEventHandler reflect.Type = reflect.TypeOf((MergeRequestEventHandler)(nil))

type eventHubCore struct {
	watch map[uint64]struct{}
	//watchProject      map[uint64][]uint64
	//watchMergeRequest map[uint64][]uint64
	listener struct {
		MergeRequest map[uint64]MergeRequestEventHandler
	}
}

func newEventHubCore() *eventHubCore {
	ctx := &eventHubCore{
		watch: make(map[uint64]struct{}),
	}
	ctx.listener.MergeRequest = make(map[uint64]MergeRequestEventHandler)
	return ctx
}

func (c *eventHubCore) Clone() *eventHubCore {
	new := newEventHubCore()
	for k, v := range c.watch {
		new.watch[k] = v
	}
	for k, v := range c.listener.MergeRequest {
		new.listener.MergeRequest[k] = v
	}
	return new
}

type eventHubExecuteContext struct {
}

func newEventHubExecuteContext() *eventHubExecuteContext {
	return &eventHubExecuteContext{}
}

func (c *eventHubExecuteContext) Clone() *eventHubExecuteContext {
	new := newEventHubExecuteContext()
	return new
}

type EventHub struct {
	Error   error
	Logger  GitlabClientLogger
	core    *eventHubCore
	context *eventHubExecuteContext

	handlerIDCounter uint64
	lock             sync.RWMutex
}

func NewEventHub() *EventHub {
	return &EventHub{
		core:    newEventHubCore(),
		context: newEventHubExecuteContext(),
	}
}

func (h *EventHub) Clone() *EventHub {
	return &EventHub{
		core:    h.core.Clone(),
		context: h.context.Clone(),
	}
}

func (h *EventHub) ContextClone() *EventHub {
	return &EventHub{
		core:    h.core,
		context: h.context.Clone(),
	}
}

func (h *EventHub) Handle(handler interface{}) error {
	tryConvert := func(target reflect.Type, value interface{}) interface{} {
		if value == nil {
			return nil
		}
		ty := reflect.TypeOf(value)
		if ty.ConvertibleTo(target) {
			return reflect.ValueOf(handler).Convert(target).Interface()
		}
		return nil
	}
	if converted := tryConvert(typeMergeRequestEventHandler, handler); converted != nil {
		return h.handleMergeRequest(converted.(MergeRequestEventHandler))
	}
	return errors.New("Unsupported handler type.")
}

func (h *EventHub) nextHandlerID() uint64 {
	return atomic.AddUint64(&h.handlerIDCounter, 1)
}

func (h *EventHub) handleMergeRequest(handler MergeRequestEventHandler) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	handlerID := h.nextHandlerID()
	h.core.listener.MergeRequest[handlerID] = handler
	h.core.watch[handlerID] = struct{}{}
	return nil
}

func (h *EventHub) Err(args ...interface{}) {
	if h.Logger == nil {
		return
	}
	h.Logger.Error(args...)
}

func (h *EventHub) Info(args ...interface{}) {
	if h.Logger == nil {
		return
	}
	h.Logger.Info(args...)
}

func (h *EventHub) ProcessWebhook(req *http.Request) (uint, error) {
	buf := bytes.Buffer{}
	bodyReader := io.TeeReader(req.Body, &buf)

	kind := &EventObjectKind{}
	if err := json.NewDecoder(bodyReader).Decode(kind); err != nil {
		h.Info("cannot decode request body: " + err.Error())
		return http.StatusBadRequest, err
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))

	switch kind.ObjectKind {
	case "merge_request":
		return h.processMergeRequest(req, &buf)
	}

	return http.StatusBadRequest, errors.New("Not supported webhook object.")
}

func (h *EventHub) processMergeRequest(req *http.Request, buf *bytes.Buffer) (uint, error) {
	var bodyReader io.Reader

	if buf == nil {
		buf = &bytes.Buffer{}
		bodyReader = io.TeeReader(req.Body, buf)
	} else {
		bodyReader = bytes.NewReader(buf.Bytes())
	}

	rawEvent := &RawMergeRequestEvent{}
	if err := json.NewDecoder(bodyReader).Decode(rawEvent); err != nil {
		h.Err("cannot decode request body as merge request event: " + err.Error())
		return http.StatusBadRequest, err
	}

	h.Info("[Gitlab Webhook] got merge request event of merge request " + strconv.FormatUint(uint64(rawEvent.ObjectAttrs.ID), 10) + ".")

	// construct event.
	// missing: Path, CreateAt, TagList, ReadmeURL, StarCount, ForkCount, LastActivityAt
	event := &MergeRequestEvent{
		User: rawEvent.User,
		SourceProject: &Project{
			ID:                rawEvent.ObjectAttrs.SourceProjectID,
			Description:       rawEvent.ObjectAttrs.Description,
			Name:              rawEvent.ObjectAttrs.Source.Name,
			PathWithNamespace: rawEvent.ObjectAttrs.Source.PathWithNamespace,
			DefaultBranch:     rawEvent.ObjectAttrs.Source.DefaultBranch,
			SSHURLToRepo:      rawEvent.ObjectAttrs.Source.GitSSHURL,
			HTTPURLToRepo:     rawEvent.ObjectAttrs.Source.GitHTTPURL,
			WebURL:            rawEvent.ObjectAttrs.Source.WebURL,
			AvatarURL:         rawEvent.ObjectAttrs.Source.AvatarURL,
			Namespace: &GitlabNamespace{
				Name: rawEvent.ObjectAttrs.Source.Namespace,
			},
		},
		TargetProject: &Project{
			ID:                rawEvent.ObjectAttrs.TargetProjectID,
			Description:       rawEvent.ObjectAttrs.Description,
			Name:              rawEvent.ObjectAttrs.Target.Name,
			PathWithNamespace: rawEvent.ObjectAttrs.Target.PathWithNamespace,
			DefaultBranch:     rawEvent.ObjectAttrs.Target.DefaultBranch,
			SSHURLToRepo:      rawEvent.ObjectAttrs.Target.GitSSHURL,
			HTTPURLToRepo:     rawEvent.ObjectAttrs.Target.GitHTTPURL,
			WebURL:            rawEvent.ObjectAttrs.Target.WebURL,
			AvatarURL:         rawEvent.ObjectAttrs.Target.AvatarURL,
			Namespace: &GitlabNamespace{
				Name: rawEvent.ObjectAttrs.Target.Namespace,
			},
		},
		MergeRequest: MergeRequest{
			ID:          rawEvent.ObjectAttrs.ID,
			InternalID:  rawEvent.ObjectAttrs.InternalID,
			ProjectID:   rawEvent.ObjectAttrs.SourceProjectID,
			Title:       rawEvent.ObjectAttrs.Title,
			Description: rawEvent.ObjectAttrs.Description,
			State:       rawEvent.ObjectAttrs.State,
			MergeBy: &User{
				ID: rawEvent.ObjectAttrs.MergeUserID,
			},
			UpdateAt:     rawEvent.ObjectAttrs.UpdateAt,
			TargetBranch: rawEvent.ObjectAttrs.TargetBranch,
			SourceBranch: rawEvent.ObjectAttrs.SourceBranch,
			Author: &Author{
				ID: rawEvent.ObjectAttrs.AuthorID,
			},
			Assignee: &Author{
				ID: rawEvent.ObjectAttrs.AssigneeID,
			},
			SourceProjectID: rawEvent.ObjectAttrs.SourceProjectID,
			TargetProjectID: rawEvent.ObjectAttrs.TargetProjectID,
			Milestone: &Milestone{
				ID: rawEvent.ObjectAttrs.MilestoneID,
			},
			MilestoneID:               rawEvent.ObjectAttrs.MilestoneID,
			MergeWhenPipelineSucceeds: rawEvent.ObjectAttrs.MergeWhenPipelineSucceeds,
			MergeStatus:               rawEvent.ObjectAttrs.MergeStatus,
			MergeCommitSHA:            rawEvent.ObjectAttrs.MergeCommitSHA,
			//ShouldRemoveSourceBranch:  rawEvent.ObjectAttrs.MergeParams.ForceRemoveSourceBranch,
		},
	}
	event.TimeStats.TimeEstimate = rawEvent.ObjectAttrs.TimeEstimate
	event.TimeStats.TotalTimeSpent = rawEvent.ObjectAttrs.TotalTimeSpent
	event.TimeStats.HumanTimeEstimate = rawEvent.ObjectAttrs.HumanTimeEstimate

	switch rawEvent.ObjectAttrs.Action {
	case "open":
		event.Event = MergeRequestOpen
	case "close":
		event.Event = MergeRequestClosed
	case "merge":
		event.Event = MergeRequestMerged
	case "reopen":
		event.Event = MergeRequestReopen
	default:
		return http.StatusOK, nil
	}

	for _, handler := range h.pickMergeRequestEventHandler(event.TargetProject.ID, event.MergeRequest.ID) {
		handler(req, event)
	}

	return http.StatusOK, nil
}

func (h *EventHub) pickMergeRequestEventHandler(projectID uint, mergeRequestID uint) (handlers []MergeRequestEventHandler) {
	handlers = make([]MergeRequestEventHandler, 0)
	appendListener := func(x uint64) {
		handler, ok := h.core.listener.MergeRequest[x]
		if !ok {
			return
		}
		handlers = append(handlers, handler)
	}
	for id, _ := range h.core.watch {
		appendListener(id)
	}
	return
}
