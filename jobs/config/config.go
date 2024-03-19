// Package config implement a simple configuration package.
// It define a type [Configuration] that represent the aplication configuration
package config

// Configuration is the configuration for the job
type Configuration struct {
	JobID    string `env:"JOB_ID,required"`
	LogLevel string `env:"LOG_LEVEL,default=INFO"`
}

type NotionalConfiguration struct {
	Environment     string `env:"ENVIRONMENT,required"`
	CoingeckoURL    string `env:"COINGECKO_URL,required"`
	CacheURL        string `env:"CACHE_URL,required"`
	CachePrefix     string `env:"CACHE_PREFIX,required"`
	NotionalChannel string `env:"NOTIONAL_CHANNEL,required"`
	P2pNetwork      string `env:"P2P_NETWORK,required"`
	AwsRegion       string `env:"AWS_REGION"`
	AwsBucket       string `env:"AWS_BUCKET"`
}

type TransferReportConfiguration struct {
	MongoURI      string `env:"MONGODB_URI,required"`
	MongoDatabase string `env:"MONGODB_DATABASE,required"`
	PageSize      int64  `env:"PAGE_SIZE,default=100"`
	PricesType    string `env:"PRICES_TYPE,required"`
	PricesUri     string `env:"PRICES_URI,required"`
	OutputPath    string `env:"OUTPUT_PATH,required"`
	P2pNetwork    string `env:"P2P_NETWORK,required"`
}

type HistoricalPricesConfiguration struct {
	MongoURI                string `env:"MONGODB_URI,required"`
	MongoDatabase           string `env:"MONGODB_DATABASE,required"`
	P2pNetwork              string `env:"P2P_NETWORK,required"`
	CoingeckoURL            string `env:"COINGECKO_URL,required"`
	CoingeckoHeaderKey      string `env:"COINGECKO_HEADER_KEY"`
	CoingeckoApiKey         string `env:"COINGECKO_API_KEY"`
	RequestLimitTimeSeconds int    `env:"REQUEST_LIMIT_TIME_SECONDS,default=5"`
	PriceDays               string `env:"PRICE_DAYS,default=max"`
}

type MigrateSourceTxConfiguration struct {
	MongoURI         string `env:"MONGODB_URI,required"`
	MongoDatabase    string `env:"MONGODB_DATABASE,required"`
	PageSize         int    `env:"PAGE_SIZE,default=100"`
	ChainID          int64  `env:"CHAIN_ID,default=0"`
	FromDate         string `env:"FROM_DATE,required"`
	ToDate           string `env:"TO_DATE,required"`
	TxTrackerURL     string `env:"TX_TRACKER_URL,required"`
	TxTrackerTimeout int64  `env:"TX_TRACKER_TIMEOUT,default=30"`
	SleepTimeSeconds int64  `env:"SLEEP_TIME_SECONDS,default=5"`
}

type ProtocolsStatsConfiguration struct {
	InfluxUrl            string     `env:"INFLUX_URL"`
	InfluxToken          string     `env:"INFLUX_TOKEN"`
	InfluxOrganization   string     `env:"INFLUX_ORGANIZATION"`
	InfluxBucket30Days   string     `env:"INFLUX_BUCKET_30_DAYS"`
	InfluxBucketInfinite string     `env:"INFLUX_BUCKET_INFINITE"`
	ProtocolsJson        string     `env:"PROTOCOLS_JSON"`
	Protocols            []Protocol `json:"PROTOCOLS"`
}

type Protocol struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type ProtocolsActivityConfiguration struct {
	ProtocolsStatsConfiguration
}
