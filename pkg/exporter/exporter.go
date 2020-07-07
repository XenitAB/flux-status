package exporter

type Exporter interface {
	Send(Event) error
}
