package resource

import (
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

func (s *ResourceUsecase) DeleteResource(ctx context.Context, id int) (*models.Resource, error) {
    baseURL := fmt.Sprintf("%s/resource", s.Config.APIBaseURL)
    params := url.Values{}
    params.Add("id", strconv.Itoa(id))
    fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

    req, err := http.NewRequestWithContext(ctx, "DELETE", fullURL, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %w", err)
    }

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

    var deleteResp models.DeleteResponse
    if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
        return nil, fmt.Errorf("error decoding response: %w", err)
    }

    s.Logger.Info("Resource deleted", zap.Int("ID", deleteResp.Data.ID))

    return &deleteResp.Data, nil
}