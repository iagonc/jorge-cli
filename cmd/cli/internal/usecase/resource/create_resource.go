package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"
	"go.uber.org/zap"
)

func (s *ResourceUsecase) CreateResource(ctx context.Context, name, dns string) (*models.Resource, error) {
    resource := models.CreateRequest{
        Name: name,
        Dns:  dns,
    }

    jsonData, err := json.Marshal(resource)
    if err != nil {
        return nil, fmt.Errorf("error marshaling JSON: %w", err)
    }

    endpoint := fmt.Sprintf("%s/resource", s.Config.APIBaseURL)
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating HTTP request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error sending request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
        return nil, utils.ParseErrorResponse(resp)
    }

    var createResp models.CreateResponse
    if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
        return nil, fmt.Errorf("error decoding response: %w", err)
    }

    s.Logger.Info("Resource created", zap.Int("ID", createResp.Data.ID))

    return &createResp.Data, nil
}