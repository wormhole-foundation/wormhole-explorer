package config

// ServiceConfiguration represents the application configuration when running as service with the default values.
type ServiceConfiguration struct {
	Environment   string `env:"ENVIRONMENT,required"`
	LogLevel      string `env:"LOG_LEVEL,default=INFO"`
	Port          string `env:"PORT,default=8000"`
	MongoURI      string `env:"MONGODB_URI,required"`
	MongoDatabase string `env:"MONGODB_DATABASE,required"`
	PprofEnabled  bool   `env:"PPROF_ENABLED,default=false"`
	P2pNetwork    string `env:"P2P_NETWORK,required"`
	AlertEnabled  bool   `env:"ALERT_ENABLED,required"`
	AlertApiKey   string `env:"ALERT_API_KEY"`

	AnkrUrl                    string `env:"ANKR_URL,required"`
	AnkrRequestsPerSecond      int    `env:"ANKR_REQUESTS_PER_SECOND,required"`
	AptosUrl                   string `env:"APTOS_URL,required"`
	AptosRequestsPerSecond     int    `env:"APTOS_REQUESTS_PER_SECOND,required"`
	ArbitrumUrl                string `env:"ARBITRUM_URL,required"`
	ArbitrumRequestsPerSecond  int    `env:"ARBITRUM_REQUESTS_PER_SECOND,required"`
	AvalancheUrl               string `env:"AVALANCHE_URL,required"`
	AvalancheRequestsPerSecond int    `env:"AVALANCHE_REQUESTS_PER_SECOND,required"`
	BaseUrl                    string `env:"BASE_URL,required"`
	BaseRequestsPerSecond      int    `env:"BASE_REQUESTS_PER_SECOND,required"`
	CeloUrl                    string `env:"CELO_URL,required"`
	CeloRequestsPerSecond      int    `env:"CELO_REQUESTS_PER_SECOND,required"`
	EthereumUrl                string `env:"ETHEREUM_URL,required"`
	EthereumRequestsPerSecond  int    `env:"ETHEREUM_REQUESTS_PER_SECOND,required"`
	MoonbeamUrl                string `env:"MOONBEAM_URL,required"`
	MoonbeamRequestsPerSecond  int    `env:"MOONBEAM_REQUESTS_PER_SECOND,required"`
	OptimismUrl                string `env:"OPTIMISM_URL,required"`
	OptimismRequestsPerSecond  int    `env:"OPTIMISM_REQUESTS_PER_SECOND,required"`
	OasisUrl                   string `env:"OASIS_URL,required"`
	OasisRequestsPerSecond     int    `env:"OASIS_REQUESTS_PER_SECOND,required"`
	PolygonUrl                 string `env:"POLYGON_URL,required"`
	PolygonRequestsPerSecond   int    `env:"POLYGON_REQUESTS_PER_SECOND,required"`
	TerraUrl                   string `env:"TERRA_URL,required"`
	TerraRequestsPerSecond     int    `env:"TERRA_REQUESTS_PER_SECOND,required"`
}

type TestnetConfiguration struct {
	BaseSepoliaBaseUrl               string `env:"BASE_SEPOLIA_URL,required"`
	BaseSepoliaRequestsPerMinute     int    `env:"BASE_SEPOLIA_REQUESTS_PER_SECOND,required"`
	EthereumSepoliaBaseUrl           string `env:"ETHEREUM_SEPOLIA_URL,required"`
	EthereumSepoliaRequestsPerMinute int    `env:"ETHEREUM_SEPOLIA_REQUESTS_PER_SECOND,required"`
}

// BackfillerConfiguration represents the application configuration when running as backfiller.
type BackfillerConfiguration struct {
	LogLevel           string `env:"LOG_LEVEL,default=INFO"`
	MongoURI           string `env:"MONGODB_URI,required"`
	MongoDatabase      string `env:"MONGODB_DATABASE,required"`
	ChainName          string `env:"CHAIN_NAME,required"`
	ChainUrl           string `env:"CHAIN_URL,required"`
	FromBlock          uint64 `env:"FROM_BLOCK,required"`
	ToBlock            uint64 `env:"TO_BLOCK,required"`
	Network            string `env:"NETWORK,required"`
	RateLimitPerSecond int    `env:"RATE_LIMIT_PER_SECOND,default=10"`
	PageSize           uint64 `env:"PAGE_SIZE,default=100"`
	PersistBlock       bool   `env:"PERSIST_BLOCK,default=false"`
}
