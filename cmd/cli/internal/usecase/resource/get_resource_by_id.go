package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"
)

func (s *ResourceUsecase) GetResourceByID(ctx context.Context, id int) (*models.Resource, error) {
    url := fmt.Sprintf("%s/resource?id=%d", s.Config.APIBaseURL, id)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

    var getResp models.GetResponse
    if err := json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
        return nil, fmt.Errorf("error decoding response: %w", err)
    }

    return &getResp.Data, nil
}