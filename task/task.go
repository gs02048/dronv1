package task

const (
	ScriptTask = 1
	RpcTaskGet = 2
	RpcTaskPost = 3
)
type Task struct {
	TaskName string `json:"name"`
	Desc string `json:"desc"`
	Command string `json:"command"`
	Args string `json:"args"`
	Path string `json:"path"`
	Spec string `json:"spec"`
	TaskType int `json:"type"`
	MaxRunTime int64 `json:"max_run_time"`
}


type TaskResult struct {
	Task
	IsSuccess int64 `json:"is_success"`
	Result string `json:"result"`
}

