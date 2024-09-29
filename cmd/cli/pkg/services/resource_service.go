package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/iagonc/jorge-cli/cmd/cli/pkg/config"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/models"
	"github.com/iagonc/jorge-cli/cmd/cli/pkg/utils"

	"go.uber.org/zap"
)

type ResourceService struct {
    Client utils.HTTPClient
    Config *config.Config
    Logger *zap.Logger
}

func NewResourceService(client utils.HTTPClient, cfg *config.Config, logger *zap.Logger) *ResourceService {
    return &ResourceService{
        Client: client,
        Config: cfg,
        Logger: logger,
    }
}

func (s *ResourceService) CreateResource(ctx context.Context, name, dns string) (*models.Resource, error) {
    // Validações
    if len(name) < 3 {
        return nil, fmt.Errorf("o nome deve ter pelo menos 3 caracteres")
    }
    if !isValidDNS(dns) {
        return nil, fmt.Errorf("formato de DNS inválido")
    }

    resource := models.CreateRequest{
        Name: name,
        Dns:  dns,
    }

    jsonData, err := json.Marshal(resource)
    if err != nil {
        return nil, fmt.Errorf("erro ao serializar JSON: %w", err)
    }

    url := fmt.Sprintf("%s/resource", s.Config.APIBaseURL)
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("erro ao criar requisição HTTP: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, utils.ParseErrorResponse(resp)
    }

    var createResp models.CreateResponse
    if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
        return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
    }

    s.Logger.Info("Recurso criado", zap.Int("ID", createResp.Data.ID))

    return &createResp.Data, nil
}

func (s *ResourceService) GetResourceByID(ctx context.Context, id int) (*models.Resource, error) {
    url := fmt.Sprintf("%s/resource?id=%d", s.Config.APIBaseURL, id)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("erro ao criar requisição: %w", err)
    }

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("recurso com ID %d não encontrado", id)
    } else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, utils.ParseErrorResponse(resp)
    }

    var getResp models.GetResponse
    if err := json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
        return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
    }

    return &getResp.Data, nil
}

func (s *ResourceService) DeleteResource(ctx context.Context, id int) (*models.Resource, error) {
    baseURL := fmt.Sprintf("%s/resource", s.Config.APIBaseURL)
    params := url.Values{}
    params.Add("id", strconv.Itoa(id))
    fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

    req, err := http.NewRequestWithContext(ctx, "DELETE", fullURL, nil)
    if err != nil {
        return nil, fmt.Errorf("erro ao criar requisição: %w", err)
    }

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("recurso com ID %d não encontrado", id)
    } else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, utils.ParseErrorResponse(resp)
    }

    var deleteResp models.DeleteResponse
    if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
        return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
    }

    s.Logger.Info("Recurso deletado", zap.Int("ID", deleteResp.Data.ID))

    return &deleteResp.Data, nil
}

func (s *ResourceService) UpdateResource(ctx context.Context, id int, name, dns string) (*models.Resource, error) {
    if name == "" && dns == "" {
        return nil, fmt.Errorf("pelo menos um de 'name' ou 'dns' deve ser fornecido")
    }

    updateReq := models.UpdateRequest{
        Name: name,
        Dns:  dns,
    }

    jsonData, err := json.Marshal(updateReq)
    if err != nil {
        return nil, fmt.Errorf("erro ao serializar JSON: %w", err)
    }

    baseURL := fmt.Sprintf("%s/resource", s.Config.APIBaseURL)
    params := url.Values{}
    params.Add("id", strconv.Itoa(id))
    fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

    req, err := http.NewRequestWithContext(ctx, "PUT", fullURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("erro ao criar requisição: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("recurso com ID %d não encontrado", id)
    } else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, utils.ParseErrorResponse(resp)
    }

    var updateResp models.UpdateResponse
    if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
        return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
    }

    s.Logger.Info("Recurso atualizado", zap.Int("ID", updateResp.Data.ID))

    return &updateResp.Data, nil
}

func (s *ResourceService) ListResources(ctx context.Context) ([]models.Resource, error) {
    url := fmt.Sprintf("%s/resources", s.Config.APIBaseURL)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("erro ao criar requisição: %w", err)
    }

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, utils.ParseErrorResponse(resp)
    }

    var apiResponse models.ApiResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
        return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
    }

    return apiResponse.Data, nil
}

func isValidDNS(dns string) bool {
    // Implementar validação de DNS conforme necessário
    return len(dns) > 3
}
