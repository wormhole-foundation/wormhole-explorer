package coingecko

import (
	"time"

	"github.com/shopspring/decimal"
)

type TokenData struct {
	ID              string `json:"id"`
	Symbol          string `json:"symbol"`
	Name            string `json:"name"`
	AssetPlatformID string `json:"asset_platform_id"`
	Platforms       struct {
		Avalanche string `json:"avalanche"`
	} `json:"platforms"`
	DetailPlatforms struct {
		Avalanche struct {
			DecimalPlace    int    `json:"decimal_place"`
			ContractAddress string `json:"contract_address"`
		} `json:"avalanche"`
	} `json:"detail_platforms"`
	BlockTimeInMinutes int      `json:"block_time_in_minutes"`
	HashingAlgorithm   any      `json:"hashing_algorithm"`
	Categories         []string `json:"categories"`
	PublicNotice       any      `json:"public_notice"`
	AdditionalNotices  []any    `json:"additional_notices"`
	Localization       struct {
		En   string `json:"en"`
		De   string `json:"de"`
		Es   string `json:"es"`
		Fr   string `json:"fr"`
		It   string `json:"it"`
		Pl   string `json:"pl"`
		Ro   string `json:"ro"`
		Hu   string `json:"hu"`
		Nl   string `json:"nl"`
		Pt   string `json:"pt"`
		Sv   string `json:"sv"`
		Vi   string `json:"vi"`
		Tr   string `json:"tr"`
		Ru   string `json:"ru"`
		Ja   string `json:"ja"`
		Zh   string `json:"zh"`
		ZhTw string `json:"zh-tw"`
		Ko   string `json:"ko"`
		Ar   string `json:"ar"`
		Th   string `json:"th"`
		ID   string `json:"id"`
		Cs   string `json:"cs"`
		Da   string `json:"da"`
		El   string `json:"el"`
		Hi   string `json:"hi"`
		No   string `json:"no"`
		Sk   string `json:"sk"`
		Uk   string `json:"uk"`
		He   string `json:"he"`
		Fi   string `json:"fi"`
		Bg   string `json:"bg"`
		Hr   string `json:"hr"`
		Lt   string `json:"lt"`
		Sl   string `json:"sl"`
	} `json:"localization"`
	Description struct {
		En   string `json:"en"`
		De   string `json:"de"`
		Es   string `json:"es"`
		Fr   string `json:"fr"`
		It   string `json:"it"`
		Pl   string `json:"pl"`
		Ro   string `json:"ro"`
		Hu   string `json:"hu"`
		Nl   string `json:"nl"`
		Pt   string `json:"pt"`
		Sv   string `json:"sv"`
		Vi   string `json:"vi"`
		Tr   string `json:"tr"`
		Ru   string `json:"ru"`
		Ja   string `json:"ja"`
		Zh   string `json:"zh"`
		ZhTw string `json:"zh-tw"`
		Ko   string `json:"ko"`
		Ar   string `json:"ar"`
		Th   string `json:"th"`
		ID   string `json:"id"`
		Cs   string `json:"cs"`
		Da   string `json:"da"`
		El   string `json:"el"`
		Hi   string `json:"hi"`
		No   string `json:"no"`
		Sk   string `json:"sk"`
		Uk   string `json:"uk"`
		He   string `json:"he"`
		Fi   string `json:"fi"`
		Bg   string `json:"bg"`
		Hr   string `json:"hr"`
		Lt   string `json:"lt"`
		Sl   string `json:"sl"`
	} `json:"description"`
	Links struct {
		Homepage                    []string `json:"homepage"`
		BlockchainSite              []string `json:"blockchain_site"`
		OfficialForumURL            []string `json:"official_forum_url"`
		ChatURL                     []string `json:"chat_url"`
		AnnouncementURL             []string `json:"announcement_url"`
		TwitterScreenName           string   `json:"twitter_screen_name"`
		FacebookUsername            string   `json:"facebook_username"`
		BitcointalkThreadIdentifier any      `json:"bitcointalk_thread_identifier"`
		TelegramChannelIdentifier   string   `json:"telegram_channel_identifier"`
		SubredditURL                any      `json:"subreddit_url"`
		ReposURL                    struct {
			Github    []any `json:"github"`
			Bitbucket []any `json:"bitbucket"`
		} `json:"repos_url"`
	} `json:"links"`
	Image struct {
		Thumb string `json:"thumb"`
		Small string `json:"small"`
		Large string `json:"large"`
	} `json:"image"`
	CountryOrigin                string  `json:"country_origin"`
	GenesisDate                  any     `json:"genesis_date"`
	ContractAddress              string  `json:"contract_address"`
	SentimentVotesUpPercentage   any     `json:"sentiment_votes_up_percentage"`
	SentimentVotesDownPercentage any     `json:"sentiment_votes_down_percentage"`
	WatchlistPortfolioUsers      int     `json:"watchlist_portfolio_users"`
	MarketCapRank                int     `json:"market_cap_rank"`
	CoingeckoRank                int     `json:"coingecko_rank"`
	CoingeckoScore               float64 `json:"coingecko_score"`
	//	DeveloperScore               int     `json:"developer_score"`
	CommunityScore      float64 `json:"community_score"`
	LiquidityScore      float64 `json:"liquidity_score"`
	PublicInterestStats struct {
		AlexaRank   any `json:"alexa_rank"`
		BingMatches any `json:"bing_matches"`
	} `json:"public_interest_stats"`
	StatusUpdates []any     `json:"status_updates"`
	LastUpdated   time.Time `json:"last_updated"`
}

type CoinHistoryResponse struct {
	Prices       [][]decimal.Decimal `json:"prices"`
	MarketCaps   [][]any             `json:"market_caps"`
	TotalVolumes [][]float64         `json:"total_volumes"`
}

type SymboleMarketCapResponse struct {
	ID         string `json"id"`
	Symbol     string `json:"symbol"`
	MarketData struct {
		MarketCap struct {
			Usd decimal.Decimal `json:"usd"`
		} `json:"market_cap"`
	} `json:"market_data"`
}
