package define

type CronLine struct {
	Name string `json:"name"`
	Command string `json:"command"`
	Spece string `json:"spece"`
	Type int32 `json:"type"`
}
