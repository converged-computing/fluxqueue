package types

type AllocationResponse struct {
	Graph Graph `json:"graph"`
}
type Paths struct {
	Containment string `json:"containment"`
}
type Node struct {
	ID       string   `json:"id"`
	Metadata Metadata `json:"metadata,omitempty"`
}
type Edges struct {
	Source string `json:"source"`
	Target string `json:"target"`
}
type Metadata struct {
	Type      string            `json:"type"`
	Id        int32             `json:"id"`
	Rank      int32             `json:"rank"`
	Basename  string            `json:"basename"`
	Exclusive bool              `json:"exclusive"`
	Paths     map[string]string `json:"paths"`
}

type Graph struct {
	Nodes []Node  `json:"nodes"`
	Edges []Edges `json:"edges"`
}
