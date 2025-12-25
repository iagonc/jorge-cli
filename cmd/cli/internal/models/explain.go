package models

// ErrorExplanation represents an explanation for an error
type ErrorExplanation struct {
	Error        string           `json:"error"`
	Category     ErrorCategory    `json:"category"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	CommonCauses []string         `json:"common_causes"`
	Solutions    []ErrorSolution  `json:"solutions"`
	Examples     []string         `json:"examples,omitempty"`
	RelatedErrors []string        `json:"related_errors,omitempty"`
	Links        []string         `json:"links,omitempty"`
}

// ErrorCategory represents the category of error
type ErrorCategory string

const (
	ErrorCategoryDNS         ErrorCategory = "DNS"
	ErrorCategoryConnection  ErrorCategory = "Connection"
	ErrorCategoryTLS         ErrorCategory = "TLS/SSL"
	ErrorCategoryHTTP        ErrorCategory = "HTTP"
	ErrorCategoryTimeout     ErrorCategory = "Timeout"
	ErrorCategoryFirewall    ErrorCategory = "Firewall"
	ErrorCategoryAuth        ErrorCategory = "Authentication"
	ErrorCategoryPermission  ErrorCategory = "Permission"
	ErrorCategoryNetwork     ErrorCategory = "Network"
	ErrorCategoryUnknown     ErrorCategory = "Unknown"
)

// ErrorSolution represents a solution for an error
type ErrorSolution struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Priority    int    `json:"priority"` // 1 = try first
}

// KnownErrors contains the database of known errors and their explanations
var KnownErrors = map[string]ErrorExplanation{
	"connection refused": {
		Error:    "connection refused",
		Category: ErrorCategoryConnection,
		Title:    "Conexão Recusada",
		Description: "O servidor recusou a conexão. Isso significa que o servidor está acessível na rede, mas nenhum serviço está escutando na porta especificada.",
		CommonCauses: []string{
			"O serviço não está rodando",
			"O serviço está rodando em uma porta diferente",
			"Firewall local bloqueando a conexão",
			"O serviço está configurado para escutar apenas em localhost",
		},
		Solutions: []ErrorSolution{
			{Title: "Verificar se o serviço está rodando", Description: "Use 'systemctl status' ou 'docker ps' para verificar", Command: "systemctl status <servico>", Priority: 1},
			{Title: "Verificar a porta correta", Description: "Confirme a porta em que o serviço está escutando", Command: "netstat -tlnp | grep <porta>", Priority: 2},
			{Title: "Verificar binding do serviço", Description: "Certifique-se que o serviço está escutando em 0.0.0.0 e não apenas 127.0.0.1", Priority: 3},
		},
		RelatedErrors: []string{"ECONNREFUSED", "connect: connection refused"},
	},
	"no such host": {
		Error:    "no such host",
		Category: ErrorCategoryDNS,
		Title:    "Host Não Encontrado",
		Description: "O nome do domínio não pôde ser resolvido para um endereço IP. O DNS não conhece esse domínio.",
		CommonCauses: []string{
			"Nome do domínio digitado incorretamente",
			"Domínio não existe ou expirou",
			"Problemas com o servidor DNS",
			"Registro DNS não propagado ainda",
		},
		Solutions: []ErrorSolution{
			{Title: "Verificar o nome do domínio", Description: "Confirme se o domínio está escrito corretamente", Priority: 1},
			{Title: "Testar com outro DNS", Description: "Tente usar o DNS do Google ou Cloudflare", Command: "nslookup <dominio> 8.8.8.8", Priority: 2},
			{Title: "Verificar /etc/hosts", Description: "Veja se há entradas locais que podem estar interferindo", Command: "cat /etc/hosts | grep <dominio>", Priority: 3},
		},
		RelatedErrors: []string{"NXDOMAIN", "getaddrinfo", "lookup"},
	},
	"timeout": {
		Error:    "timeout",
		Category: ErrorCategoryTimeout,
		Title:    "Tempo Limite Excedido",
		Description: "A operação demorou mais do que o tempo permitido. O servidor pode estar lento, sobrecarregado ou inacessível.",
		CommonCauses: []string{
			"Servidor sobrecarregado",
			"Rede lenta ou congestionada",
			"Firewall bloqueando silenciosamente (drop)",
			"Rota de rede incorreta",
			"Timeout configurado muito baixo",
		},
		Solutions: []ErrorSolution{
			{Title: "Aumentar o timeout", Description: "Tente aumentar o tempo limite da operação", Priority: 1},
			{Title: "Verificar conectividade básica", Description: "Teste se o host responde a ping", Command: "ping -c 3 <host>", Priority: 2},
			{Title: "Verificar rota de rede", Description: "Analise o caminho até o destino", Command: "traceroute <host>", Priority: 3},
			{Title: "Verificar firewall", Description: "Confirme que não há regras de firewall bloqueando", Priority: 4},
		},
		RelatedErrors: []string{"ETIMEDOUT", "context deadline exceeded", "i/o timeout"},
	},
	"certificate": {
		Error:    "certificate",
		Category: ErrorCategoryTLS,
		Title:    "Erro de Certificado SSL/TLS",
		Description: "Há um problema com o certificado SSL/TLS do servidor. Pode estar expirado, ser inválido ou não confiável.",
		CommonCauses: []string{
			"Certificado expirado",
			"Certificado auto-assinado",
			"Nome do host não corresponde ao certificado",
			"Cadeia de certificados incompleta",
			"Certificado revogado",
		},
		Solutions: []ErrorSolution{
			{Title: "Verificar data de expiração", Description: "Confirme se o certificado não expirou", Command: "echo | openssl s_client -connect <host>:443 2>/dev/null | openssl x509 -noout -dates", Priority: 1},
			{Title: "Verificar o certificado", Description: "Analise detalhes do certificado", Command: "openssl s_client -connect <host>:443 -showcerts", Priority: 2},
			{Title: "Renovar certificado", Description: "Use Let's Encrypt ou seu provedor de certificados para renovar", Priority: 3},
		},
		RelatedErrors: []string{"x509", "certificate has expired", "certificate signed by unknown authority"},
	},
	"permission denied": {
		Error:    "permission denied",
		Category: ErrorCategoryPermission,
		Title:    "Permissão Negada",
		Description: "Você não tem permissão para executar esta operação. Pode ser um problema de permissão de arquivo, rede ou sistema.",
		CommonCauses: []string{
			"Porta privilegiada (< 1024) requer root",
			"Arquivo ou diretório sem permissão de leitura/escrita",
			"SELinux ou AppArmor bloqueando",
			"Permissões de rede restritas",
		},
		Solutions: []ErrorSolution{
			{Title: "Usar sudo", Description: "Execute o comando com privilégios elevados", Command: "sudo <comando>", Priority: 1},
			{Title: "Verificar permissões", Description: "Verifique as permissões do arquivo ou recurso", Command: "ls -la <arquivo>", Priority: 2},
			{Title: "Verificar SELinux", Description: "Veja se o SELinux está bloqueando", Command: "getenforce && ausearch -m avc -ts recent", Priority: 3},
		},
		RelatedErrors: []string{"EACCES", "EPERM", "access denied"},
	},
	"network unreachable": {
		Error:    "network unreachable",
		Category: ErrorCategoryNetwork,
		Title:    "Rede Inacessível",
		Description: "Não há rota de rede para o destino. O sistema não sabe como alcançar o endereço de destino.",
		CommonCauses: []string{
			"Sem conexão de rede",
			"Gateway/roteador desconfigurado",
			"Rota de rede ausente",
			"Interface de rede desativada",
		},
		Solutions: []ErrorSolution{
			{Title: "Verificar interface de rede", Description: "Confirme que a interface está ativa", Command: "ip link show", Priority: 1},
			{Title: "Verificar gateway", Description: "Confirme que há um gateway configurado", Command: "ip route show", Priority: 2},
			{Title: "Reiniciar rede", Description: "Tente reiniciar o serviço de rede", Command: "sudo systemctl restart NetworkManager", Priority: 3},
		},
		RelatedErrors: []string{"ENETUNREACH", "no route to host"},
	},
	"too many open files": {
		Error:    "too many open files",
		Category: ErrorCategoryPermission,
		Title:    "Muitos Arquivos Abertos",
		Description: "O processo atingiu o limite de arquivos abertos. Isso inclui sockets de rede.",
		CommonCauses: []string{
			"Limite de ulimit muito baixo",
			"Vazamento de conexões/arquivos no código",
			"Muitas conexões simultâneas",
		},
		Solutions: []ErrorSolution{
			{Title: "Aumentar limite temporário", Description: "Aumente o limite de arquivos abertos", Command: "ulimit -n 65535", Priority: 1},
			{Title: "Verificar limite atual", Description: "Veja quantos arquivos podem ser abertos", Command: "ulimit -n", Priority: 2},
			{Title: "Aumentar limite permanente", Description: "Edite /etc/security/limits.conf", Priority: 3},
		},
		RelatedErrors: []string{"EMFILE", "ENFILE"},
	},
	"connection reset": {
		Error:    "connection reset",
		Category: ErrorCategoryConnection,
		Title:    "Conexão Resetada",
		Description: "A conexão foi abruptamente fechada pelo servidor ou por um dispositivo intermediário.",
		CommonCauses: []string{
			"Servidor reiniciou ou crashou",
			"Firewall ou WAF bloqueou a conexão",
			"Load balancer timeout",
			"Configuração de keep-alive incorreta",
		},
		Solutions: []ErrorSolution{
			{Title: "Verificar logs do servidor", Description: "Analise os logs para entender o motivo do reset", Priority: 1},
			{Title: "Verificar firewall/WAF", Description: "Confirme que não há bloqueios", Priority: 2},
			{Title: "Tentar novamente", Description: "Pode ser um problema temporário", Priority: 3},
		},
		RelatedErrors: []string{"ECONNRESET", "connection reset by peer"},
	},
}
