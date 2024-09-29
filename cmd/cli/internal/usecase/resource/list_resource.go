package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"
)

func (s *ResourceUsecase) ListResources(ctx context.Context) ([]models.Resource, error) {
    url := fmt.Sprintf("%s/resources", s.Config.APIBaseURL)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %w", err)
    }

    resp, err := s.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error sending request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, utils.ParseErrorResponse(resp)
    }

    var apiResponse models.ApiResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
        return nil, fmt.Errorf("error decoding response: %w", err)
    }

    return apiResponse.Data, nil
}