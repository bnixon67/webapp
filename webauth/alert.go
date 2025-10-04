package webauth

type AlertType string

const (
	AlertInfo  AlertType = "info"
	AlertWarn  AlertType = "warn"
	AlertError AlertType = "error"
)

type Alert struct {
	Type AlertType
	Text string
}
