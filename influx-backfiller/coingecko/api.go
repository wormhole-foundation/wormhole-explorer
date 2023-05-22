package coingecko

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CoinGeckoAPI struct {
	ApiURL     string
	ApiKey     string
	tokenCache map[string]TokenItem
}

type TokenItem struct {
	Id       string
	Chain    string
	Symbol   string
	Decimals int
}

func NewCoinGeckoAPI(apiKey string) *CoinGeckoAPI {
	return &CoinGeckoAPI{
		ApiURL:     "api.coingecko.com",
		ApiKey:     apiKey,
		tokenCache: make(map[string]TokenItem),
	}
}

func (cg *CoinGeckoAPI) GetSymbolDailyPrice(CoinID string) (*CoinHistoryResponse, error) {

	url := fmt.Sprintf("https://%s/api/v3/coins/%s/market_chart?vs_currency=usd&days=max&interval=daily", cg.ApiURL, CoinID)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, fmt.Errorf("token not found")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var td CoinHistoryResponse
	err = json.Unmarshal(body, &td)
	if err != nil {
		return nil, err
	}

	return &td, nil

}

func (cg *CoinGeckoAPI) GetSymbol(ChainId string, ContractId string) (string, error) {

	// lookup on cache first
	if val, ok := cg.tokenCache[ContractId]; ok {
		return val.Symbol, nil
	}

	// lookup on coingecko
	ti, err := cg.GetSymbolByContract(ChainId, ContractId)
	if err != nil {

		// if not found, return none
		if err.Error() == "token not found" {
			return "none", nil
		}

		return "", err

	}

	// add to cache
	fmt.Printf("adding to cache: %s\n", ti.Symbol)
	cg.tokenCache[ContractId] = *ti
	return ti.Symbol, nil

}

func (cg *CoinGeckoAPI) convertChain(chain string) string {

	// check if chain exsist on map
	if val, ok := convertionMap[chain]; ok {
		return val
	}
	return chain

}

// GetSymbolByContract returns the symbol of the token
// Input: ChaindId is the name of the chain: ie: ethereum, solana, etc
// Input: ContractId is the contract address of the token (ECR-20 or other)
func (cg *CoinGeckoAPI) GetSymbolByContract(ChainId string, ContractId string) (*TokenItem, error) {

	chain := cg.convertChain(ChainId)
	//url := "https://api.coingecko.com/api/v3/coins/avalanche/contract/0x2b2c81e08f1af8835a78bb2a90ae924ace0ea4be"
	url := fmt.Sprintf("https://%s/api/v3/coins/%s/contract/%s", cg.ApiURL, chain, ContractId)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	//req.Header.Add("Cookie", "__cf_bm=jUWxA1U8U3SdvDF2EXgCZUmnDopOozWnB5VpXIjWH.c-1682970763-0-AaLD4yVrSy53aAJQwVNe61P5IcXSnW4vIMeRrsRDIMGJ/+PbEcOv/lene34+FB4Q4kapT//4660lx/Rw507zw7Q=")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, fmt.Errorf("token not found")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var td TokenData
	err = json.Unmarshal(body, &td)
	if err != nil {
		return nil, err
	}

	ti := TokenItem{
		Id:     td.ID,
		Chain:  ChainId,
		Symbol: td.Symbol,
	}

	fmt.Printf("\"%s\": \"%s\",\n", ContractId, ti.Symbol)

	return &ti, nil
}
