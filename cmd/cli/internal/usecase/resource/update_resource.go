package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"
	"go.uber.org/zap"
)

func (s *ResourceUsecase) UpdateResource(ctx context.Context, id int, name, dns string) (*models.Resource, error) {
    if name == "" && dns == "" {
        return nil, fmt.Errorf("at least one of 'name' or 'dns' must be provided")
    }

    updateReq := models.UpdateRequest{
        Name: name,
        Dns:  dns,
    }

    jsonData, err := json.Marshal(updateReq)
    if err != nil {
        return nil, fmt.Errorf("error marshaling JSON: %w", err)
    }

    baseURL := fmt.Sprintf("%s/resource", s.Config.APIBaseURL)
    params := url.Values{}
    params.Add("id", strconv.Itoa(id))
    fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

    req, err := http.NewRequestWithContext(ctx, "PUT", fullURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("error creating request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error sending request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("resource with ID %d not found", id)
    } else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, utils.ParseErrorResponse(resp)
    }

    var updateResp models.UpdateResponse
    if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
        return nil, fmt.Errorf("error decoding response: %w", err)
    }

    s.Logger.Info("Resource updated", zap.Int("ID", updateResp.Data.ID))

    return &updateResp.Data, nil
}