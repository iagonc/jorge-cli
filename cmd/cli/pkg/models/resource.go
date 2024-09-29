package models

type Resource struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Dns       string `json:"dns"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type CreateRequest struct {
    Name string `json:"name"`
    Dns  string `json:"dns"`
}

type CreateResponse struct {
    Data    Resource `json:"data"`
    Message string   `json:"message"`
}

type UpdateRequest struct {
    Name string `json:"name,omitempty"`
    Dns  string `json:"dns,omitempty"`
}

type UpdateResponse struct {
    Data    Resource `json:"data"`
    Message string   `json:"message"`
}

type DeleteResponse struct {
    Data    Resource `json:"data"`
    Message string   `json:"message"`
}

type ApiResponse struct {
    Data    []Resource `json:"data"`
    Message string     `json:"message"`
}
