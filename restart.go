package app

import "sync/atomic"

type RestartPolicy string

func (r RestartPolicy) String() string {
	return string(r)
}

const (
	RestartNever     RestartPolicy = "never"
	RestartAlways    RestartPolicy = "always"
	RestartOnError   RestartPolicy = "on_error"
	RestartOnSuccess RestartPolicy = "on_success"
)

type restartHandler struct {
	policy         RestartPolicy
	restartCounter atomic.Uint64
	maxRestarts    uint64
}

func NewRestartHandler(restart RestartPolicy, maxRestarts uint64) *restartHandler {
	return &restartHandler{
		policy:      restart,
		maxRestarts: maxRestarts,
	}
}

func (h *restartHandler) IncrementRestartCounter() {
	h.restartCounter.Add(1)
}

func (h *restartHandler) GetRestartCounter() uint64 {
	return h.restartCounter.Load()
}

func (h *restartHandler) GetMaxRestarts() uint64 {
	return h.maxRestarts
}

func (h *restartHandler) ShouldRestart(err error) bool {
	return h.shouldRestart(err) && h.hasRestartQuota()
}

func (h *restartHandler) shouldRestart(err error) bool {
	switch h.policy {
	case RestartNever:
		return false
	case RestartAlways:
		return true
	case RestartOnError:
		return err != nil
	case RestartOnSuccess:
		return err == nil
	default:
		return false
	}
}

func (h *restartHandler) hasRestartQuota() bool {
	return h.maxRestarts == 0 || h.restartCounter.Load() < h.maxRestarts
}
