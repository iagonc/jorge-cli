package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iagonc/jorge-cli/internal/handler"
	"github.com/iagonc/jorge-cli/internal/repository"
	"github.com/iagonc/jorge-cli/internal/schemas"
	"github.com/iagonc/jorge-cli/internal/usecase"
	"github.com/iagonc/jorge-cli/test/testutil"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db := testutil.SetupTestDB(t)
	logger := testutil.NewTestLogger()
	repo := repository.NewSQLiteResourceRepository(db)

	createUseCase := usecase.NewCreateResource(repo, logger)
	deleteUseCase := usecase.NewDeleteResource(repo, logger)
	getResourceByIDUseCase := usecase.NewGetResourceByID(repo, logger)
	listUseCase := usecase.NewListResources(repo, logger)
	listByNameUseCase := usecase.NewListResourcesByName(repo, logger)
	updateUseCase := usecase.NewUpdateResource(repo, logger)

	h := handler.NewHandler(
		createUseCase,
		deleteUseCase,
		getResourceByIDUseCase,
		listUseCase,
		listByNameUseCase,
		updateUseCase,
		logger,
	)

	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/resources", h.ListResourcesHandler)
		v1.GET("/resources/name", h.ListResourcesByNameHandler)
		v1.GET("/resource", h.GetResourceByIDHandler)
		v1.POST("/resource", h.CreateResourceHandler)
		v1.PUT("/resource", h.UpdateResourceHandler)
		v1.DELETE("/resource", h.DeleteResourceHandler)
	}

	return router, db
}

func TestListResources_Integration(t *testing.T) {
	router, db := setupTestRouter(t)
	testutil.SeedTestData(t, db)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/resources", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(data) != 3 {
		t.Errorf("Expected 3 resources, got %d", len(data))
	}
}

func TestCreateResource_Integration(t *testing.T) {
	router, _ := setupTestRouter(t)

	resource := schemas.Resource{
		Name: "integration-test",
		Dns:  "integration.example.com",
	}
	jsonBody, _ := json.Marshal(resource)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/resource", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestResourceCRUD_Integration(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Create
	createBody, _ := json.Marshal(schemas.Resource{
		Name: "crud-test",
		Dns:  "crud.example.com",
	})
	createReq, _ := http.NewRequest(http.MethodPost, "/api/v1/resource", bytes.NewBuffer(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusOK {
		t.Fatalf("Create failed: %s", createW.Body.String())
	}

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	resourceID := int(data["ID"].(float64))

	// Read
	getReq, _ := http.NewRequest(http.MethodGet, "/api/v1/resource?id="+string(rune(resourceID+'0')), nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	// Delete
	deleteReq, _ := http.NewRequest(http.MethodDelete, "/api/v1/resource?id="+string(rune(resourceID+'0')), nil)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)
}
