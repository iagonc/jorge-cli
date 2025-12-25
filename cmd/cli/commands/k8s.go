package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/usecase/k8s"
	"github.com/spf13/cobra"
)

// NewK8sCommand creates the k8s command
func NewK8sCommand(usecase *k8s.K8sUsecase) *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:     "k8s",
		Aliases: []string{"kube", "kubectl"},
		Short:   "Ferramentas Kubernetes",
		Long: `Ferramentas para interagir com clusters Kubernetes.

Subcomandos:
  ctx      - Gerenciar contexts
  pods     - Listar pods
  deploy   - Listar deployments
  svc      - Listar services
  logs     - Ver logs de um pod
  events   - Ver eventos
  restart  - Reiniciar deployment
  scale    - Escalar deployment`,
		Example: `  # Listar contexts
  jorge k8s ctx

  # Listar pods
  jorge k8s pods -n default

  # Ver logs
  jorge k8s logs my-pod -n default

  # Reiniciar deployment
  jorge k8s restart my-deploy -n default`,
	}

	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace (default: default)")

	// Context subcommand
	ctxCmd := &cobra.Command{
		Use:   "ctx [name]",
		Short: "Gerenciar contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Switch context
				if err := usecase.UseContext(args[0]); err != nil {
					return err
				}
				fmt.Printf("\n   %s Switched to context: %s\n\n",
					lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
					args[0])
				return nil
			}

			// List contexts
			contexts, err := usecase.GetContexts()
			if err != nil {
				return err
			}

			displayK8sContexts(contexts)
			return nil
		},
	}

	// Pods subcommand
	podsCmd := &cobra.Command{
		Use:   "pods",
		Short: "Listar pods",
		RunE: func(cmd *cobra.Command, args []string) error {
			pods, err := usecase.GetPods(namespace)
			if err != nil {
				return err
			}

			displayK8sResources("Pods", pods)
			return nil
		},
	}

	// Deployments subcommand
	deployCmd := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"deployments"},
		Short:   "Listar deployments",
		RunE: func(cmd *cobra.Command, args []string) error {
			deploys, err := usecase.GetDeployments(namespace)
			if err != nil {
				return err
			}

			displayK8sResources("Deployments", deploys)
			return nil
		},
	}

	// Services subcommand
	svcCmd := &cobra.Command{
		Use:     "svc",
		Aliases: []string{"services"},
		Short:   "Listar services",
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := usecase.GetServices(namespace)
			if err != nil {
				return err
			}

			displayK8sResources("Services", services)
			return nil
		},
	}

	// Logs subcommand
	logsCmd := &cobra.Command{
		Use:   "logs <pod>",
		Short: "Ver logs de um pod",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			lines, _ := cmd.Flags().GetInt("tail")

			logs, err := usecase.GetPodLogs(namespace, args[0], lines, false)
			if err != nil {
				return err
			}

			fmt.Println(logs)
			return nil
		},
	}
	logsCmd.Flags().Int("tail", 100, "Numero de linhas")

	// Describe subcommand
	describeCmd := &cobra.Command{
		Use:   "describe <pod>",
		Short: "Descrever pod",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := usecase.DescribePod(namespace, args[0])
			if err != nil {
				return err
			}

			fmt.Println(output)
			return nil
		},
	}

	// Events subcommand
	eventsCmd := &cobra.Command{
		Use:   "events",
		Short: "Ver eventos",
		RunE: func(cmd *cobra.Command, args []string) error {
			events, err := usecase.GetEvents(namespace)
			if err != nil {
				return err
			}

			displayK8sEvents(events)
			return nil
		},
	}

	// Restart subcommand
	restartCmd := &cobra.Command{
		Use:   "restart <deployment>",
		Short: "Reiniciar deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := usecase.RestartDeployment(namespace, args[0]); err != nil {
				return err
			}

			fmt.Printf("\n   %s Deployment %s reiniciado\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				args[0])
			return nil
		},
	}

	// Scale subcommand
	scaleCmd := &cobra.Command{
		Use:   "scale <deployment> <replicas>",
		Short: "Escalar deployment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var replicas int
			fmt.Sscanf(args[1], "%d", &replicas)

			if err := usecase.ScaleDeployment(namespace, args[0], replicas); err != nil {
				return err
			}

			fmt.Printf("\n   %s Deployment %s escalado para %d replicas\n\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
				args[0], replicas)
			return nil
		},
	}

	// Port-forward subcommand
	portForwardCmd := &cobra.Command{
		Use:   "port-forward <pod> <local:remote>",
		Short: "Port forward para um pod",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var local, remote int
			fmt.Sscanf(args[1], "%d:%d", &local, &remote)

			fmt.Printf("\n   Forwarding localhost:%d -> %s:%d\n", local, args[0], remote)
			fmt.Println("   Press Ctrl+C to stop")

			return usecase.PortForward(namespace, args[0], local, remote)
		},
	}

	// Apply subcommand
	applyCmd := &cobra.Command{
		Use:   "apply <file>",
		Short: "Aplicar manifesto",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := usecase.ApplyManifest(args[0])
			if err != nil {
				return err
			}

			fmt.Println(output)
			return nil
		},
	}

	// Generate subcommand
	genCmd := &cobra.Command{
		Use:   "gen <kind> <name>",
		Short: "Gerar manifesto",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			image, _ := cmd.Flags().GetString("image")

			manifest, err := usecase.GenerateManifest(args[0], args[1], image)
			if err != nil {
				return err
			}

			fmt.Println(manifest)
			return nil
		},
	}
	genCmd.Flags().String("image", "nginx:latest", "Imagem do container")

	cmd.AddCommand(ctxCmd, podsCmd, deployCmd, svcCmd, logsCmd, describeCmd, eventsCmd, restartCmd, scaleCmd, portForwardCmd, applyCmd, genCmd)

	return cmd
}

func displayK8sContexts(contexts []models.K8sContext) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("☸ Kubernetes Contexts"))
	fmt.Println()

	for _, ctx := range contexts {
		marker := "  "
		style := lipgloss.NewStyle()
		if ctx.Current {
			marker = "* "
			style = style.Foreground(lipgloss.Color("42")).Bold(true)
		}
		fmt.Printf("   %s%s\n", marker, style.Render(ctx.Name))
	}
	fmt.Println()
}

func displayK8sResources(kind string, resources []models.K8sResource) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("☸ " + kind))
	fmt.Println()

	if len(resources) == 0 {
		fmt.Println("   Nenhum recurso encontrado")
		fmt.Println()
		return
	}

	// Dynamic header based on resource type
	fmt.Println(strings.Repeat("─", 90))

	switch resources[0].Kind {
	case "Pod":
		fmt.Printf("   %-35s %-12s %-8s %-10s %-10s %s\n",
			lipgloss.NewStyle().Bold(true).Render("NAME"),
			lipgloss.NewStyle().Bold(true).Render("STATUS"),
			lipgloss.NewStyle().Bold(true).Render("READY"),
			lipgloss.NewStyle().Bold(true).Render("RESTARTS"),
			lipgloss.NewStyle().Bold(true).Render("AGE"),
			lipgloss.NewStyle().Bold(true).Render("NODE"))
		fmt.Println(strings.Repeat("─", 90))

		for _, r := range resources {
			statusStyle := getPodStatusStyle(r.Status)
			fmt.Printf("   %-35s %-12s %-8s %-10d %-10s %s\n",
				truncateK8sStr(r.Name, 33),
				statusStyle.Render(r.Status),
				r.Ready,
				r.Restarts,
				r.Age,
				truncateK8sStr(r.Node, 20))
		}

	case "Deployment":
		fmt.Printf("   %-35s %-12s %-10s %s\n",
			lipgloss.NewStyle().Bold(true).Render("NAME"),
			lipgloss.NewStyle().Bold(true).Render("READY"),
			lipgloss.NewStyle().Bold(true).Render("REPLICAS"),
			lipgloss.NewStyle().Bold(true).Render("AGE"))
		fmt.Println(strings.Repeat("─", 90))

		for _, r := range resources {
			fmt.Printf("   %-35s %-12s %-10s %s\n",
				truncateK8sStr(r.Name, 33),
				r.Ready,
				r.Replicas,
				r.Age)
		}

	case "Service":
		fmt.Printf("   %-30s %-12s %-18s %s\n",
			lipgloss.NewStyle().Bold(true).Render("NAME"),
			lipgloss.NewStyle().Bold(true).Render("TYPE"),
			lipgloss.NewStyle().Bold(true).Render("CLUSTER-IP"),
			lipgloss.NewStyle().Bold(true).Render("AGE"))
		fmt.Println(strings.Repeat("─", 90))

		for _, r := range resources {
			fmt.Printf("   %-30s %-12s %-18s %s\n",
				truncateK8sStr(r.Name, 28),
				r.Status,
				r.IP,
				r.Age)
		}
	}

	fmt.Println(strings.Repeat("─", 90))
	fmt.Println()
}

func displayK8sEvents(events []models.K8sEvent) {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("☸ Events"))
	fmt.Println()

	if len(events) == 0 {
		fmt.Println("   Nenhum evento encontrado")
		fmt.Println()
		return
	}

	for _, e := range events {
		typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		if e.Type == "Warning" {
			typeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		}

		fmt.Printf("   %s [%s] %s (x%d)\n",
			typeStyle.Render(e.Type),
			e.Reason,
			truncateK8sStr(e.Message, 60),
			e.Count)
	}
	fmt.Println()
}

func getPodStatusStyle(status string) lipgloss.Style {
	switch status {
	case "Running":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	case "Pending":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case "Failed", "Error", "CrashLoopBackOff":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	case "Completed", "Succeeded":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}

func truncateK8sStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
