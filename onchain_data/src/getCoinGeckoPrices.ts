import axios from "axios";
import { COIN_GECKO_EXCEPTIONS } from "./utils";
const CoinGecko = require("coingecko-api");

function sleep(milliseconds) {
  const date = Date.now();
  let currentDate = 0;
  do {
    currentDate = Date.now();
  } while (currentDate - date < milliseconds);
}

//2. Initiate the CoinGecko API Client
const CoinGeckoClient = new CoinGecko(); //use justin's VIP coin gecko api key?

async function getCoinsList() {
  let data = await CoinGeckoClient.coins.list();
  //console.log(data)
  return;
}

export async function getTokenPrice(tokenAddress) {
  let tokenContractInfo = await CoinGeckoClient.coins.fetchCoinContractInfo(
    tokenAddress
  );
  let data = tokenContractInfo["data"];
  var price = 0;

  try {
    price = data["market_data"]["current_price"].usd;
  } catch (e) {
    console.log("could not find price for address=", tokenAddress);
  }
  return price;
}

export async function getTokenPrices(tokenAddresses) {
  //only for ethereum chain
  let tokenContractInfos;
  try {
    tokenContractInfos = await CoinGeckoClient.simple.fetchTokenPrice({
      contract_addresses: tokenAddresses,
      vs_currencies: "usd",
    });
  } catch (e) {
    console.log(e);
    console.log("could not find prices for addresses");
  }
  return tokenContractInfos["data"];
}

function getKeyByValue(object, value) {
  return Object.keys(object).find((key) => object[key] === value);
}

export async function getCoinGeckoMap() {
  // pull coin gecko mapping of token address to coingecko id
  var coinMap;
  const map_query = `https://api.coingecko.com/api/v3/coins/list?include_platform=true`;
  try {
    const req = await axios.get(map_query);
    coinMap = req.data;
  } catch (e) {
    console.log(e);
    console.log("could not find prices for addresses");
  }
  var coinMapTransformed: any[] = [];
  for (let i = 0; i < coinMap.length; i++) {
    const coin = coinMap[i];

    const platforms = coin?.platforms;
    Object.entries(platforms).forEach((entry) => {
      const [platform, contractAddress] = entry;
      if (contractAddress != "" && contractAddress != null) {
        coinMapTransformed.push({
          contractAddress: contractAddress,
          coinGeckoId: coin.id,
        });
      }
    });
  }
  return coinMapTransformed;
}

export async function getTokenPricesGET(
  chainId: number,
  platform: String,
  tokenAddresses: String[]
) {
  let data;

  const addresses = tokenAddresses.join("%2C");
  const price_query = `https://api.coingecko.com/api/v3/simple/token_price/${platform}?contract_addresses=${addresses}&vs_currencies=usd`;
  //console.log(price_query)
  try {
    const tokenContractInfos = await axios.get(price_query);
    data = tokenContractInfos.data;
  } catch (e) {
    console.log(e);
    console.log("could not find prices for addresses");
  }

  // find prices included in exceptions using coin gecko id and add it back to the results
  const filteredIds = COIN_GECKO_EXCEPTIONS.filter(
    (x) => tokenAddresses.includes(x.tokenAddress) && x.chainId == chainId
  );
  const coinGeckoIds = filteredIds.map((x) => x.coinGeckoId);

  const additional_data = await getTokenPricesCGID(coinGeckoIds);
  let gecko_id_data = {};
  for (let i = 0; i < filteredIds.length; i++) {
    let cg_id = filteredIds[i];
    gecko_id_data[cg_id.tokenAddress] = additional_data[cg_id.coinGeckoId];
  }

  return { ...data, ...gecko_id_data };
}

export async function getTokenPricesCGID(coinGeckoIds: String[]) {
  let data;
  const ids = coinGeckoIds.join("%2C");
  const price_query = `https://api.coingecko.com/api/v3/simple/price?ids=${ids}&vs_currencies=usd`;
  console.log(price_query);
  try {
    const tokenContractInfos = await axios.get(price_query);
    data = tokenContractInfos.data;
  } catch (e) {
    // console.log(e);
    console.log("could not find prices for ids", coinGeckoIds);
  }
  return data;
}
