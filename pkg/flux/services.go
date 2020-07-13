package flux

type Service struct {
	Id        string
	ReadOnly  string
	Status    string
	Rollout   Rollout
	SyncError string
	Automated bool
	Locked    bool
	Ignore    bool
}

type Rollout struct {
	Desired   int
	Updated   int
	Ready     int
	Available int
	Outdated  int
}
