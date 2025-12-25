package testutil

import (
	"net/http"

	"github.com/iagonc/jorge-cli/internal/schemas"
)

// MockHTTPClient implements the HTTPClient interface for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// MockResourceRepository implements the ResourceRepository interface for testing
type MockResourceRepository struct {
	CreateFunc      func(resource *schemas.Resource) error
	DeleteFunc      func(id uint) error
	FindByNameFunc  func(name string) (*schemas.Resource, error)
	FindByDNSFunc   func(dns string) (*schemas.Resource, error)
	FindByIDFunc    func(id uint) (*schemas.Resource, error)
	ListFunc        func() ([]*schemas.Resource, error)
	UpdateFunc      func(resource *schemas.Resource) error
	ListByNameFunc  func(name string) ([]*schemas.Resource, error)
}

func (m *MockResourceRepository) Create(resource *schemas.Resource) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(resource)
	}
	return nil
}

func (m *MockResourceRepository) Delete(id uint) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockResourceRepository) FindByName(name string) (*schemas.Resource, error) {
	if m.FindByNameFunc != nil {
		return m.FindByNameFunc(name)
	}
	return nil, nil
}

func (m *MockResourceRepository) FindByDNS(dns string) (*schemas.Resource, error) {
	if m.FindByDNSFunc != nil {
		return m.FindByDNSFunc(dns)
	}
	return nil, nil
}

func (m *MockResourceRepository) FindByID(id uint) (*schemas.Resource, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, nil
}

func (m *MockResourceRepository) List() ([]*schemas.Resource, error) {
	if m.ListFunc != nil {
		return m.ListFunc()
	}
	return nil, nil
}

func (m *MockResourceRepository) Update(resource *schemas.Resource) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(resource)
	}
	return nil
}

func (m *MockResourceRepository) ListByName(name string) ([]*schemas.Resource, error) {
	if m.ListByNameFunc != nil {
		return m.ListByNameFunc(name)
	}
	return nil, nil
}

// MockCommandExecutor implements command execution for testing network tools
type MockCommandExecutor struct {
	ExecuteFunc func(name string, args ...string) ([]byte, error)
}

func (m *MockCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(name, args...)
	}
	return nil, nil
}

// MockTLSDialer implements TLS dialing for testing SSL checks
type MockTLSDialer struct {
	DialFunc func(network, addr string) (interface{}, error)
}
