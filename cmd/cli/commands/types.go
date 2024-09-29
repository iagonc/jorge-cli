package commands

type Resource struct {
    ID   uint   `json:"ID"`
    Name string `json:"name"`
    Dns  string `json:"dns"`
}
