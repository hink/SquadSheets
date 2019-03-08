package models

// AdminRole administrator role
type AdminRole struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type User struct {
	Name    string `json:"name"`
	Steam64 string `json:"steam64"`
	Role    string `json:"role.omitempty"`
	Notes   string `json:"notes,omitempty"`
}
