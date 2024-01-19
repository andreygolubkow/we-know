package historical_code_storage

type HistoricalCodeStorage interface {
	SetUp() error
	Update() error

	Cleanup() error
}
