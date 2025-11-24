package robot

// ExecutionInput é o contrato de entrada do seu robô.
// Ele serve tanto para o JSON da API quanto para argumentos de CLI.
type ExecutionInput struct {
	// Auth: Flexível para qualquer tipo de login (user/pass, token, client_id/secret)
	Auth map[string]string `json:"auth"`

	// Params: Dados variáveis da execução (filtros, datas, IDs)
	Params map[string]any `json:"params"`
}

// Helper para validar se uma credencial existe
func (i *ExecutionInput) GetCredential(key string) string {
	if val, ok := i.Auth[key]; ok {
		return val
	}
	return ""
}

func (i *ExecutionInput) GetParams(key string) any {
	if val, ok := i.Params[key]; ok {
		return val
	}
	return ""
}