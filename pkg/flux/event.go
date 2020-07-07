package flux

type Event struct {
	Id        int      `json:"id"`
	Type      string   `json:"type"`
	StartedAt string   `json:"startedAt"`
	EndedAt   string   `json:"endedAt"`
	Time      string   `json:"time"`
	Metadata  Metadata `json:"metadata"`
}

type Metadata struct {
	Commits []Commit `json:"commits"`
	Errors  []Error  `json:"errors"`
}

type Commit struct {
	Revision string `json:"revision"`
}

type Error struct {
	Id    string `json:"ID"`
	Path  string `json:"Path"`
	Error string `json:"Error"`
}
