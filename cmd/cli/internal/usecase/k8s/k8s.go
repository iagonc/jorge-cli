package k8s

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// K8sUsecase handles Kubernetes operations
type K8sUsecase struct {
	logger     *zap.Logger
	kubeconfig string
}

// NewK8sUsecase creates a new K8sUsecase
func NewK8sUsecase(logger *zap.Logger) *K8sUsecase {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	return &K8sUsecase{
		logger:     logger,
		kubeconfig: kubeconfig,
	}
}

// GetContexts lists available Kubernetes contexts
func (u *K8sUsecase) GetContexts() ([]models.K8sContext, error) {
	if !u.hasKubectl() {
		return nil, fmt.Errorf("kubectl not found in PATH")
	}

	// Get current context
	currentCtx, _ := u.runKubectl("config", "current-context")
	currentCtx = strings.TrimSpace(currentCtx)

	// Get all contexts
	output, err := u.runKubectl("config", "get-contexts", "-o", "name")
	if err != nil {
		return nil, err
	}

	var contexts []models.K8sContext
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			continue
		}
		contexts = append(contexts, models.K8sContext{
			Name:    name,
			Current: name == currentCtx,
		})
	}

	return contexts, nil
}

// UseContext switches to a different context
func (u *K8sUsecase) UseContext(name string) error {
	_, err := u.runKubectl("config", "use-context", name)
	return err
}

// GetPods lists pods in a namespace
func (u *K8sUsecase) GetPods(namespace string) ([]models.K8sResource, error) {
	if namespace == "" {
		namespace = "default"
	}

	output, err := u.runKubectl("get", "pods", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}

	var podList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
				ContainerStatuses []struct {
					Ready        bool `json:"ready"`
					RestartCount int  `json:"restartCount"`
				} `json:"containerStatuses"`
				PodIP string `json:"podIP"`
			} `json:"status"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &podList); err != nil {
		return nil, err
	}

	var pods []models.K8sResource
	for _, item := range podList.Items {
		ready := 0
		total := len(item.Status.ContainerStatuses)
		restarts := 0
		for _, cs := range item.Status.ContainerStatuses {
			if cs.Ready {
				ready++
			}
			restarts += cs.RestartCount
		}

		pods = append(pods, models.K8sResource{
			Kind:      "Pod",
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Status:    item.Status.Phase,
			Ready:     fmt.Sprintf("%d/%d", ready, total),
			Restarts:  restarts,
			Age:       formatAge(item.Metadata.CreationTimestamp),
			Node:      item.Spec.NodeName,
			IP:        item.Status.PodIP,
		})
	}

	return pods, nil
}

// GetDeployments lists deployments in a namespace
func (u *K8sUsecase) GetDeployments(namespace string) ([]models.K8sResource, error) {
	if namespace == "" {
		namespace = "default"
	}

	output, err := u.runKubectl("get", "deployments", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}

	var deployList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Replicas          int `json:"replicas"`
				ReadyReplicas     int `json:"readyReplicas"`
				AvailableReplicas int `json:"availableReplicas"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &deployList); err != nil {
		return nil, err
	}

	var deployments []models.K8sResource
	for _, item := range deployList.Items {
		deployments = append(deployments, models.K8sResource{
			Kind:      "Deployment",
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Ready:     fmt.Sprintf("%d/%d", item.Status.ReadyReplicas, item.Status.Replicas),
			Replicas:  fmt.Sprintf("%d", item.Status.Replicas),
			Age:       formatAge(item.Metadata.CreationTimestamp),
		})
	}

	return deployments, nil
}

// GetServices lists services in a namespace
func (u *K8sUsecase) GetServices(namespace string) ([]models.K8sResource, error) {
	if namespace == "" {
		namespace = "default"
	}

	output, err := u.runKubectl("get", "services", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}

	var svcList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
			Spec struct {
				Type      string `json:"type"`
				ClusterIP string `json:"clusterIP"`
				Ports     []struct {
					Port       int    `json:"port"`
					TargetPort interface{} `json:"targetPort"`
				} `json:"ports"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &svcList); err != nil {
		return nil, err
	}

	var services []models.K8sResource
	for _, item := range svcList.Items {
		ports := ""
		for i, p := range item.Spec.Ports {
			if i > 0 {
				ports += ","
			}
			ports += fmt.Sprintf("%d", p.Port)
		}

		services = append(services, models.K8sResource{
			Kind:      "Service",
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Status:    item.Spec.Type,
			IP:        item.Spec.ClusterIP,
			Age:       formatAge(item.Metadata.CreationTimestamp),
		})
	}

	return services, nil
}

// GetPodLogs gets logs from a pod
func (u *K8sUsecase) GetPodLogs(namespace, podName string, lines int, follow bool) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	args := []string{"logs", "-n", namespace, podName}
	if lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", lines))
	}

	return u.runKubectl(args...)
}

// DescribePod describes a pod
func (u *K8sUsecase) DescribePod(namespace, podName string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	return u.runKubectl("describe", "pod", "-n", namespace, podName)
}

// GetEvents gets events in a namespace
func (u *K8sUsecase) GetEvents(namespace string) ([]models.K8sEvent, error) {
	if namespace == "" {
		namespace = "default"
	}

	output, err := u.runKubectl("get", "events", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}

	var eventList struct {
		Items []struct {
			Type    string `json:"type"`
			Reason  string `json:"reason"`
			Message string `json:"message"`
			Count   int    `json:"count"`
			FirstTimestamp time.Time `json:"firstTimestamp"`
			LastTimestamp  time.Time `json:"lastTimestamp"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &eventList); err != nil {
		return nil, err
	}

	var events []models.K8sEvent
	for _, item := range eventList.Items {
		events = append(events, models.K8sEvent{
			Type:      item.Type,
			Reason:    item.Reason,
			Message:   item.Message,
			Count:     item.Count,
			FirstSeen: item.FirstTimestamp,
			LastSeen:  item.LastTimestamp,
		})
	}

	return events, nil
}

// RestartDeployment restarts a deployment
func (u *K8sUsecase) RestartDeployment(namespace, name string) error {
	if namespace == "" {
		namespace = "default"
	}

	_, err := u.runKubectl("rollout", "restart", "deployment", name, "-n", namespace)
	return err
}

// ScaleDeployment scales a deployment
func (u *K8sUsecase) ScaleDeployment(namespace, name string, replicas int) error {
	if namespace == "" {
		namespace = "default"
	}

	_, err := u.runKubectl("scale", "deployment", name, "-n", namespace, "--replicas", fmt.Sprintf("%d", replicas))
	return err
}

// PortForward sets up port forwarding
func (u *K8sUsecase) PortForward(namespace, podName string, localPort, remotePort int) error {
	if namespace == "" {
		namespace = "default"
	}

	cmd := exec.Command("kubectl", "port-forward", "-n", namespace, podName,
		fmt.Sprintf("%d:%d", localPort, remotePort))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ApplyManifest applies a YAML manifest
func (u *K8sUsecase) ApplyManifest(path string) (string, error) {
	return u.runKubectl("apply", "-f", path)
}

// DeleteResource deletes a resource
func (u *K8sUsecase) DeleteResource(kind, namespace, name string) error {
	if namespace == "" {
		namespace = "default"
	}

	_, err := u.runKubectl("delete", kind, name, "-n", namespace)
	return err
}

func (u *K8sUsecase) hasKubectl() bool {
	_, err := exec.LookPath("kubectl")
	return err == nil
}

func (u *K8sUsecase) runKubectl(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, string(output))
	}
	return string(output), nil
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	if d.Hours() >= 24 {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

// GenerateManifest generates a sample Kubernetes manifest
func (u *K8sUsecase) GenerateManifest(kind, name, image string) (string, error) {
	var manifest interface{}

	switch kind {
	case "deployment":
		manifest = map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]string{
				"name": name,
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]string{"app": name},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]string{"app": name},
					},
					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  name,
								"image": image,
								"ports": []map[string]int{{"containerPort": 80}},
							},
						},
					},
				},
			},
		}

	case "service":
		manifest = map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]string{
				"name": name,
			},
			"spec": map[string]interface{}{
				"selector": map[string]string{"app": name},
				"ports": []map[string]interface{}{
					{"port": 80, "targetPort": 80},
				},
				"type": "ClusterIP",
			},
		}

	default:
		return "", fmt.Errorf("unsupported kind: %s", kind)
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
