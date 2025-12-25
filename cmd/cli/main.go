package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/iagonc/jorge-cli/cmd/cli/commands"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/config"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/audit"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/bench"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/chaos"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/compare"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/convert"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/cronx"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/curl"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/database"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/deps"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/diagnose"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/diff"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/dns"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/encode"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/explain"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/headers"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/health"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/incident"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/ipinfo"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/jsonq"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/jwt"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/k8s"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/logs"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/mock"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/mtr"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/network"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/notify"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/playbook"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/portscan"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/postmortem"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/redis"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/resource"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/secrets"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/slo"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/ssl"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/status"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/timex"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/trace"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/tunnel"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/watch"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"
)

func main() {
	// Initialize the logger
	logger, err := utils.InitializeLogger()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configurations
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Error loading configurations", zap.Error(err))
		os.Exit(1)
	}

	// Initialize the HTTP client
	client := utils.NewHTTPClient(cfg.Timeout)

	// Initialize usecases
	resourceUsecase := resource.NewResourceUsecase(client, cfg, logger)
	networkUsecase := network.NewNetworkDebugUsecase(logger)

	// Initialize new troubleshooting usecases
	sslUsecase := ssl.NewSSLCheckUsecase(logger)
	portscanUsecase := portscan.NewPortScanUsecase(logger)
	healthUsecase := health.NewHealthCheckUsecase(logger)
	benchUsecase := bench.NewHTTPBenchmarkUsecase(logger)
	dnsUsecase := dns.NewDNSDiffUsecase(logger)
	playbookUsecase := playbook.NewPlaybookUsecase(logger)

	// Initialize diagnose usecases
	diagnoseUsecase := diagnose.NewDiagnoseUsecase(logger)
	doctorUsecase := diagnose.NewDoctorUsecase(logger)

	// Initialize new advanced troubleshooting usecases
	traceUsecase := trace.NewTraceUsecase(logger)
	watchUsecase := watch.NewWatchUsecase(logger)
	compareUsecase := compare.NewCompareUsecase(logger)
	depsUsecase := deps.NewDepsUsecase(logger)
	mtrUsecase := mtr.NewMTRUsecase(logger)
	headersUsecase := headers.NewHeadersUsecase(logger)
	curlUsecase := curl.NewCurlUsecase(logger)
	mockUsecase := mock.NewMockUsecase(logger)
	tunnelUsecase := tunnel.NewTunnelUsecase(logger)
	explainUsecase := explain.NewExplainUsecase(logger)
	auditUsecase := audit.NewAuditUsecase(logger)

	// Batch 1: Utilities
	encodeUsecase := encode.NewEncodeUsecase(logger)
	cronUsecase := cronx.NewCronUsecase(logger)
	timeUsecase := timex.NewTimeUsecase(logger)
	jwtUsecase := jwt.NewJWTUsecase(logger)

	// Batch 2: Data Processing
	jsonqUsecase := jsonq.NewJSONUsecase(logger)
	convertUsecase := convert.NewConvertUsecase(logger)
	diffUsecase := diff.NewDiffUsecase(logger)
	logsUsecase := logs.NewLogsUsecase(logger)

	// Batch 3: Security
	ipUsecase := ipinfo.NewIPUsecase(logger)
	secretsUsecase := secrets.NewSecretsUsecase(logger)

	// Batch 4: Infrastructure
	redisUsecase := redis.NewRedisUsecase(logger)
	databaseUsecase := database.NewDatabaseUsecase(logger)
	statusUsecase := status.NewStatusUsecase(logger)

	// Batch 5: Incident
	incidentUsecase := incident.NewIncidentUsecase(logger)
	notifyUsecase := notify.NewNotifyUsecase(logger)
	postmortemUsecase := postmortem.NewPostmortemUsecase(logger)

	// Batch 6: Advanced
	k8sUsecase := k8s.NewK8sUsecase(logger)
	sloUsecase := slo.NewSLOUsecase(logger)
	chaosUsecase := chaos.NewChaosUsecase(logger)

	// Set up the root command
	var rootCmd = &cobra.Command{
		Use:     "jorge-cli",
		Short:   "Jorge CLI - A friendly network diagnostic and resource management tool",
		Long:    "A command-line tool to perform network diagnostics, troubleshooting, and manage resources via API.",
		Version: cfg.Version,
	}

	// Add resource management commands
	rootCmd.AddCommand(commands.NewListCommand(resourceUsecase))
	rootCmd.AddCommand(commands.NewCreateCommand(resourceUsecase))
	rootCmd.AddCommand(commands.NewDeleteCommand(resourceUsecase))
	rootCmd.AddCommand(commands.NewUpdateCommand(resourceUsecase))

	// Add network diagnostic commands
	rootCmd.AddCommand(commands.NewNetworkDebugCommand(networkUsecase))

	// Add new troubleshooting commands
	rootCmd.AddCommand(commands.NewSSLCheckCommand(sslUsecase))
	rootCmd.AddCommand(commands.NewPortScanCommand(portscanUsecase))
	rootCmd.AddCommand(commands.NewHealthCheckCommand(healthUsecase))
	rootCmd.AddCommand(commands.NewHTTPBenchCommand(benchUsecase))
	rootCmd.AddCommand(commands.NewDNSDiffCommand(dnsUsecase))

	// Add playbook command
	rootCmd.AddCommand(commands.NewPlaybookCommand(playbookUsecase))

	// Add diagnose and doctor commands (main troubleshooting commands)
	rootCmd.AddCommand(commands.NewDiagnoseCommand(diagnoseUsecase))
	rootCmd.AddCommand(commands.NewDoctorCommand(doctorUsecase))

	// Add advanced troubleshooting commands
	rootCmd.AddCommand(commands.NewTraceCommand(traceUsecase))
	rootCmd.AddCommand(commands.NewWatchCommand(watchUsecase))
	rootCmd.AddCommand(commands.NewCompareCommand(compareUsecase))
	rootCmd.AddCommand(commands.NewDepsCommand(depsUsecase))
	rootCmd.AddCommand(commands.NewMTRCommand(mtrUsecase))
	rootCmd.AddCommand(commands.NewHeadersCommand(headersUsecase))
	rootCmd.AddCommand(commands.NewCurlCommand(curlUsecase))
	rootCmd.AddCommand(commands.NewMockCommand(mockUsecase))
	rootCmd.AddCommand(commands.NewTunnelCommand(tunnelUsecase))
	rootCmd.AddCommand(commands.NewExplainCommand(explainUsecase))
	rootCmd.AddCommand(commands.NewAuditCommand(auditUsecase))

	// Batch 1: Utilities commands
	rootCmd.AddCommand(commands.NewEncodeCommand(encodeUsecase))
	rootCmd.AddCommand(commands.NewCronCommand(cronUsecase))
	rootCmd.AddCommand(commands.NewTimeCommand(timeUsecase))
	rootCmd.AddCommand(commands.NewJWTCommand(jwtUsecase))

	// Batch 2: Data Processing commands
	rootCmd.AddCommand(commands.NewJSONCommand(jsonqUsecase))
	rootCmd.AddCommand(commands.NewConvertCommand(convertUsecase))
	rootCmd.AddCommand(commands.NewDiffCommand(diffUsecase))
	rootCmd.AddCommand(commands.NewLogsCommand(logsUsecase))

	// Batch 3: Security commands
	rootCmd.AddCommand(commands.NewIPCommand(ipUsecase))
	rootCmd.AddCommand(commands.NewSecretsCommand(secretsUsecase))

	// Batch 4: Infrastructure commands
	rootCmd.AddCommand(commands.NewRedisCommand(redisUsecase))
	rootCmd.AddCommand(commands.NewDatabaseCommand(databaseUsecase))
	rootCmd.AddCommand(commands.NewStatusCommand(statusUsecase))

	// Batch 5: Incident commands
	rootCmd.AddCommand(commands.NewIncidentCommand(incidentUsecase))
	rootCmd.AddCommand(commands.NewNotifyCommand(notifyUsecase))
	rootCmd.AddCommand(commands.NewPostmortemCommand(postmortemUsecase))

	// Batch 6: Advanced commands
	rootCmd.AddCommand(commands.NewK8sCommand(k8sUsecase))
	rootCmd.AddCommand(commands.NewSLOCommand(sloUsecase))
	rootCmd.AddCommand(commands.NewChaosCommand(chaosUsecase))

	// Handle system signals for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go utils.HandleSignals(cancel, logger)

	// Execute the root command with context
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.Error("Error executing command", zap.Error(err))
		os.Exit(1)
	}
}
