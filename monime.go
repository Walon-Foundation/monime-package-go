package monime

type MonimeVersion string

const (
	Version20250823 MonimeVersion = "caph.2025-08-23"
	Version20250620 MonimeVersion = "caph.2025-06-20"
)

// ClientOptions is what the user fills out
type ClientOptions struct {
	AccessToken string
	SpaceID     string
	Version     *MonimeVersion
}

// Client is the main container
type Client struct {
	// Private config
	accessToken string
	spaceID     string
	version     *MonimeVersion

	// ... add other services here
	
}

func New(opts ClientOptions) *Client {
	// 1. Initialize the main client with the options
	c := &Client{
		accessToken: opts.AccessToken,
		spaceID:     opts.SpaceID,
		version:     opts.Version,
	}

	
	return c
}


