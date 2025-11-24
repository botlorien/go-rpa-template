package botapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type Config struct {
	APIURL   string
	User     string
	Password string
}

type Client struct {
	Config      Config // Guarda a config
	BotInstance *Bot
	BotName     string
	HTTPClient  *http.Client
}

// NewClient inicializa o cliente lendo variáveis de ambiente
func NewClient(cfg Config) (*Client, error) {
    if cfg.APIURL == "" {
        return nil, fmt.Errorf("URL da API do BotApp é obrigatória")
    }

    // CORREÇÃO: Faça o Trim na variável 'cfg' ANTES de criar o Client
    cfg.APIURL = strings.TrimRight(cfg.APIURL, "/")

    return &Client{
        Config:     cfg, // Aqui você passa a config já ajustada
        HTTPClient: &http.Client{Timeout: 30 * time.Second},
    }, nil
}

// SetBot registra ou atualiza o Bot na API
func (c *Client) SetBot(name, description, version, department string) error {
	// Limpeza do nome (Regex portado do Python)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
	cleanedName := reg.ReplaceAllString(name, " ")
	cleanedName = strings.Title(strings.TrimSpace(cleanedName))
	c.BotName = cleanedName

	payload := map[string]interface{}{
		"name":        cleanedName,
		"description": description,
		"version":     version,
		"department":  strings.ToUpper(department),
		"is_active":   true,
	}

	// 1. Tenta buscar o bot
	existingBot, err := c.searchBot(cleanedName)
	if err != nil {
		return err
	}

	if existingBot != nil {
		c.BotInstance = existingBot
		// Lógica de Patch se mudou algo
		if existingBot.Description != description || existingBot.Version != version || existingBot.Department != strings.ToUpper(department) {
			_, err := c.doRequest("PATCH", fmt.Sprintf("/bots/%d/", existingBot.ID), payload)
			if err != nil {
				return fmt.Errorf("falha ao atualizar bot: %v", err)
			}
			// Atualiza instância local
			c.BotInstance.Description = description
			c.BotInstance.Version = version
			c.BotInstance.Department = strings.ToUpper(department)
		}
	} else {
		// Create
		resp, err := c.doRequest("POST", "/bots/", payload)
		if err != nil {
			return fmt.Errorf("falha ao criar bot: %v", err)
		}
		var newBot Bot
		if err := json.Unmarshal(resp, &newBot); err != nil {
			return err
		}
		c.BotInstance = &newBot
	}

	return nil
}

// RunTask é o wrapper (o "decorator") que envolve sua função
// funcName: Nome da tarefa na dashboard
// description: Descrição da tarefa
// taskFunc: A função que contém sua lógica
func (c *Client) RunTask(funcName, description string, taskFunc func() (any, error)) (any, error) {
	if c.BotInstance == nil {
		return nil, fmt.Errorf("bot não definido. Chame SetBot() antes")
	}
	if !c.BotInstance.IsActive {
		return nil, fmt.Errorf("o bot '%s' está inativo", c.BotInstance.Name)
	}

	// 1. Registra/Busca a Task na API
	taskObj, err := c.ensureTask(funcName, description)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar task: %v", err)
	}

	// 2. Coleta dados do ambiente
	envInfo := c.collectEnvInfo()
	startTime := time.Now()

	// 3. Cria Log (STARTED)
	logPayload := envInfo
	logPayload.TaskID = taskObj.ID
	logPayload.Status = "started"
	logPayload.StartTime = startTime
	
	logResp, err := c.doRequest("POST", "/tasklog/", logPayload)
	if err != nil {
		fmt.Printf("⚠️ Erro ao criar log de início: %v\n", err)
	}
	
	// Extrai ID do log criado para atualizar depois
	var createdLog struct { ID int `json:"id"` }
	_ = json.Unmarshal(logResp, &createdLog)
	logID := createdLog.ID

	// 4. Executa a função do usuário
	var result any
	var execErr error
	var panicErr interface{}
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = r // Captura panic (crash)
			}
		}()
		result, execErr = taskFunc()
	}()

	endTime := time.Now()
	duration := endTime.Sub(startTime).String()

	// 5. Prepara o Payload de Finalização (PATCH)
	finalPayload := map[string]interface{}{
		"end_time": endTime,
		"duration": duration,
	}

	if panicErr != nil {
		// Se deu Panic (Crash)
		finalPayload["status"] = "failed"
		finalPayload["error_message"] = fmt.Sprintf("PANIC: %v\nStack: %s", panicErr, string(debug.Stack()))
		finalPayload["exception_type"] = "Panic"
	} else if execErr != nil {
		// Se a função retornou erro
		finalPayload["status"] = "failed"
		finalPayload["error_message"] = execErr.Error()
		finalPayload["exception_type"] = "Error"
	} else {
		// Sucesso
		finalPayload["status"] = "completed"
		finalPayload["result_data"] = map[string]string{"return": fmt.Sprintf("%v", result)}
	}

	// 6. Atualiza o Log
	if logID != 0 {
		_, err := c.doRequest("PATCH", fmt.Sprintf("/tasklog/%d/", logID), finalPayload)
		if err != nil {
			fmt.Printf("⚠️ Erro ao fechar log: %v\n", err)
		}
	}

	// Retorna o resultado original para o código chamador
	if panicErr != nil {
		panic(panicErr) // Relança o pânico localmente
	}
	return result, execErr
}

// --- Métodos Privados Auxiliares ---

func (c *Client) searchBot(name string) (*Bot, error) {
	resp, err := c.doRequest("GET", "/bots/?search="+name, nil)
	if err != nil {
		return nil, err
	}
	var bots []Bot
	if err := json.Unmarshal(resp, &bots); err != nil {
		return nil, err
	}
	for _, b := range bots {
		if b.Name == name {
			return &b, nil
		}
	}
	return nil, nil
}

func (c *Client) ensureTask(name, description string) (*Task, error) {
	// Busca tasks existentes
	resp, err := c.doRequest("GET", fmt.Sprintf("/tasks/?bot=%d&name=%s", c.BotInstance.ID, name), nil)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	if err := json.Unmarshal(resp, &tasks); err != nil {
		return nil, err
	}

	for _, t := range tasks {
		if t.Name == name {
			// Update description if needed
			if t.Description != description {
				c.doRequest("PATCH", fmt.Sprintf("/tasks/%d/", t.ID), map[string]string{"description": description})
			}
			return &t, nil
		}
	}

	// Create
	newPayload := map[string]interface{}{
		"bot":         c.BotInstance.ID,
		"name":        name,
		"description": description,
	}
	resp, err = c.doRequest("POST", "/tasks/", newPayload)
	if err != nil {
		return nil, err
	}
	var newTask Task
	json.Unmarshal(resp, &newTask)
	return &newTask, nil
}

func (c *Client) collectEnvInfo() LogPayload {
	hostname, _ := os.Hostname()
	currentUser, _ := user.Current()
	wd, _ := os.Getwd()
	
	username := "unknown"
	if currentUser != nil {
		username = currentUser.Username
	}

	env := os.Getenv("BOTAPP_DEPLOY_ENV")
	if env == "" {
		env = "dev"
	}

	return LogPayload{
		HostName:      hostname,
		HostIP:        "0.0.0.0", // Go requer pacotes externos para IP real confiável, deixei placeholder
		UserLogin:     username,
		BotDir:        wd,
		OSPlatform:    runtime.GOOS + "/" + runtime.GOARCH,
		PythonVersion: runtime.Version(), // Na verdade é Go Version
		PID:           os.Getpid(),
		Env:           env,
		TriggerSource: "cli",
		ManualTrigger: true,
	}
}

func (c *Client) doRequest(method, endpoint string, data interface{}) ([]byte, error) {
	url := c.BaseURL + endpoint
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.User, c.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return respBody, fmt.Errorf("API Error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, err
}