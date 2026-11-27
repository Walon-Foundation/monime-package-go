package monime

// MonimeVersion identifies an API version sent via the Monime-Version header.
type MonimeVersion string

const (
	Version20250823 MonimeVersion = "caph.2025-08-23"
	Version20250620 MonimeVersion = "caph.2025-06-20"

	// DefaultVersion is sent when neither WithVersion nor MONIME_VERSION
	// chooses one. It is the latest release.
	DefaultVersion = Version20250823
)

// Amount is a monetary value expressed in minor units (e.g. cents).
type Amount struct {
	Currency string `json:"currency"`
	Value    int64  `json:"value"`
}

// Pagination holds the paging metadata returned by list endpoints.
type Pagination struct {
	Count int `json:"count"`

	// Deprecated: the API does not return a total; pagination is
	// {count, next} only. Total is always zero and kept so existing code still
	// compiles. It will be removed in a later release.
	Total int `json:"total"`

	// Next is the cursor for the following page; pass it to WithAfter. It is
	// empty on the last page.
	Next string `json:"next"`
}
