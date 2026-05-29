package execution

type OrderResult struct {
	Success bool

	Venue  string
	Symbol string

	OrderID string
	Status  string

	Message string

	Raw any
}