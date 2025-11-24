package botapp

import "time"

// Bot representa a estrutura do robô na API
type Bot struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Department  string `json:"department"`
	IsActive    bool   `json:"is_active"`
}

// Task representa a tarefa registrada
type Task struct {
	ID          int    `json:"id"`
	BotID       int    `json:"bot"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// LogPayload representa o payload para criar/atualizar logs
type LogPayload struct {
	TaskID        int       `json:"task,omitempty"`
	Status        string    `json:"status"`
	StartTime     time.Time `json:"start_time,omitempty"`
	EndTime       time.Time `json:"end_time,omitempty"`
	Duration      string    `json:"duration,omitempty"` // API espera string no formato de tempo do Django/Python
	HostIP        string    `json:"host_ip"`
	HostName      string    `json:"host_name"`
	UserLogin     string    `json:"user_login"`
	BotDir        string    `json:"bot_dir"`
	OSPlatform    string    `json:"os_platform"`
	PythonVersion string    `json:"python_version"` // Usaremos GoVersion aqui
	PID           int       `json:"pid"`
	Env           string    `json:"env"`
	TriggerSource string    `json:"trigger_source"`
	ManualTrigger bool      `json:"manual_trigger"`
	ResultData    any       `json:"result_data,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	ExceptionType string    `json:"exception_type,omitempty"`
}