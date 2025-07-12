package report

import (
	hs "we-know/pkg/infrastructure/historical_code_storage"
)

// Reporter defines the interface for generating reports
type Reporter interface {
	// GenerateReport generates a report for the given code storage
	// Returns the path to the generated report and an error if any
	GenerateReport(codeStorage hs.HistoricalCodeStorage) (string, error)
}
