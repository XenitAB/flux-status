package exporter

import "errors"

type Exporter interface {
	Send(Event) error
	String() string
}

func GetExporter(url string, azdoPat string, gitlabToken string) (Exporter, error) {
	gitlab, err := NewGitlab(gitlabToken, url)
	if err == nil {
		return gitlab, nil
	}

	azdo, err := NewAzureDevops(azdoPat, url)
	if err == nil {
		return azdo, nil
	}

	return nil, errors.New("Could not find a compatible exporter")
}
