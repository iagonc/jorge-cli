package commands

type Resource struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Dns       string `json:"dns"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}
