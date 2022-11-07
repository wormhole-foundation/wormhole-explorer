import { LCDClient } from "@terra-money/terra.js";
import { formatUnits } from "ethers/lib/utils";

import { getTokenPricesCGID, getTokenPricesGET } from "./getCoinGeckoPrices";

import {
  CHAIN_ID_TERRA,
  CHAIN_ID_TERRA2,
  isNativeCosmWasmDenom,
  getEmitterAddressTerra,
} from "@certusone/wormhole-sdk";

import axios from "axios";
import { DISALLOWLISTED_ADDRESSES, CHAIN_INFO_MAP } from "./utils";

// current allowlist used for stats/govenor
import allowList = require("./allowList.json");
import { BigNumber } from "ethers";

require("dotenv").config();

function formatTerraNative(address: string) {
  let symbol = address.slice(1);
  if (address != "uluna") {
    symbol = symbol.slice(0, -1) + "t";
  }
  return symbol;
}

function getAllowList(chainId) {
  if (Object.keys(allowList).includes(chainId.toString())) {
    return allowList[chainId];
  } else {
    return [];
  }
}

function calcTokenQty(tokenInfo) {
  return Number(formatUnits(tokenInfo.balance, tokenInfo.decimals));
}

async function contractQuery(
  address: string,
  query: object,
  host
): Promise<any> {
  //return this.provider.wasm.contractQuery(address, query);
  const lcd = new LCDClient(host);
  let info = undefined;
  try {
    info = await lcd.wasm.contractQuery(address, query);
  } catch (e) {
    console.log("could not query contract info: ", address);
  }

  return info;
}

async function queryTokenInfo(address: string, host): Promise<any> {
  return await contractQuery(
    address,
    {
      token_info: {},
    },
    host
  );
}

function getNativeTerraTokenInfo(address, chainId) {
  if (address === "uluna") {
    if (chainId === CHAIN_ID_TERRA) {
      return {
        name: "Luna Classic",
        symbol: "LUNC",
        decimals: 6,
        tokenAddress: address,
      };
    } else {
      return {
        name: "Luna",
        symbol: "LUNA",
        decimals: 6,
        tokenAddress: address,
      };
    }
  } else if (address === "uusd") {
    return {
      name: "UST",
      symbol: "UST",
      decimals: 6,
      tokenAddress: address,
    };
  }
  return {};
}

export async function getTerraTokenAccounts(chainInfo, useAllowList) {
  const chainId = chainInfo.chain_id;
  const bridgeAddress = chainInfo.token_bridge_address;

  var TERRA_HOST;
  var network;
  if (chainId == CHAIN_ID_TERRA) {
    TERRA_HOST = {
      URL: "https://columbus-lcd.terra.dev",
      chainID: "columbus-5",
      name: "mainnet",
    };
    network = "classic";
  } else if (chainId == CHAIN_ID_TERRA2) {
    TERRA_HOST = {
      URL: "https://phoenix-lcd.terra.dev",
      chainID: "phoenix-1",
      name: "mainnet",
    };
    network = "mainnet";
  }
  const lcd = new LCDClient(TERRA_HOST);

  var tokenList = [];
  var tokenData = {};
  if (useAllowList) {
    const allowList = getAllowList(chainId);
    Object.keys(allowList).forEach((address) => {
      tokenList.push(address);
    });
    for (let i = 0; i < tokenList.length; i++) {
      let address = tokenList[i];
      var tokenInfo = undefined;
      if (isNativeCosmWasmDenom(chainId, address)) {
        console.log("isNativeCosmWasmDenom?", address);
        tokenInfo = getNativeTerraTokenInfo(address, chainId);
      } else {
        tokenInfo = await queryTokenInfo(address, TERRA_HOST);
      }

      tokenData[address] = tokenInfo;
    }
    // console.log("tokenData", tokenData);
  } else {
    const token_url = "https://assets.terra.money/cw20/tokens.json";
    const response = await axios.get(token_url);
    if (response.status == 200) {
      const data = response["data"];

      if (data.hasOwnProperty(network)) {
        tokenData = data[network];
        tokenList = Object.keys(tokenData);
      } else {
        console.log("object has no data");
      }
    } else {
      console.log(response.status);
    }
  }

  let tokenAccounts = [];
  let nativeTokens = [];
  try {
    const address = bridgeAddress;
  } catch (e) {
    console.log(e);
  }
  for (let i = 0; i < tokenList.length; i++) {
    const tokenAddress = tokenList[i];

    try {
      let token = await lcd.wasm.contractQuery(tokenAddress, {
        balance: { address: bridgeAddress },
      });
      let tokenInfo = tokenData[tokenAddress];
      // console.log(tokenAddress, token, tokenInfo);

      tokenAccounts.push({
        tokenAddress: tokenAddress,
        name: tokenInfo.name,
        decimals: tokenInfo?.decimals || 6,
        symbol: tokenInfo.symbol,
        balance: BigNumber.from(token["balance"]),
      });
    } catch (e) {
      console.log("could not find address=", tokenAddress);
    }
  }
  // grab native tokens
  const [balance] = await lcd.bank.balance(bridgeAddress);
  balance.toData().forEach((token) => {
    if (useAllowList) {
      if (tokenList.includes(token.denom)) {
        let tokenInfo = tokenData[token.denom];
        tokenAccounts.push({
          tokenAddress: token.denom,
          name: tokenInfo.name,
          decimals: tokenInfo?.decimals || 6,
          symbol: tokenInfo.symbol,
          balance: BigNumber.from(token.amount),
        });
      }
    } else {
      tokenAccounts.push({
        tokenAddress: token.denom,
        name: "native",
        symbol: formatTerraNative(token.denom),
        balance: token.amount,
        decimals: "6",
      });
    }
  });
  // console.log("tokenAccounts=", tokenAccounts);

  return tokenAccounts;
}

async function getTokenValues(chainInfo, tokenInfos: any[], useAllowList) {
  try {
    const custody = tokenInfos.map((tokenInfo) => ({
      ...tokenInfo,
      qty: calcTokenQty(tokenInfo),
    }));

    const custodyFiltered = custody.filter((c) => c.qty > 0);
    var tokenPrices = {};
    var prices = [];

    if (useAllowList) {
      // use coingecko ids from allowlist
      const allowList = getAllowList(chainInfo.chain_id);
      const cgids: string[] = Object.values(allowList);

      // input array of cgids, returns json with cgid:price
      prices = await getTokenPricesCGID(cgids);
      if (prices === undefined) {
        console.log(`could not find ids for ${chainInfo.chain_id}`);
        return [];
      }
      for (const [key, value] of Object.entries(prices)) {
        if (!value.hasOwnProperty("usd")) {
          prices[key] = { usd: 0 };
        }
      }
      // have to map cgid: price to tokenAddress: price
      for (const [key, value] of Object.entries(allowList)) {
        for (const [key1, value1] of Object.entries(prices)) {
          if (key1 === value) {
            tokenPrices[key] = prices[key1];
          }
        }
      }
    } else {
      let j = 0;
      let chunk_size = 100;
      while (j < custodyFiltered.length) {
        prices = await getTokenPricesGET(
          chainInfo.chain_id,
          chainInfo.platform,
          custodyFiltered.slice(j, j + chunk_size).map((x) => x.tokenAddress)
        );

        for (const [key, value] of Object.entries(prices)) {
          if (!value.hasOwnProperty("usd")) {
            prices[key] = { usd: 0 };
          }
        }

        tokenPrices = { ...tokenPrices, ...prices };
        j += chunk_size;
      }
    }

    // filter list by those with coin gecko prices
    const filteredBalances = custodyFiltered.filter((x) =>
      Object.keys(tokenPrices).includes(x?.tokenAddress)
    );
    // calculate usd balances. add price and usd balance to tokenInfos
    const balancesUSD = filteredBalances.map((tokenInfo) => ({
      ...tokenInfo,
      tokenPrice: tokenPrices[tokenInfo?.tokenAddress]["usd"],
      tokenBalanceUSD:
        tokenInfo.qty * tokenPrices[tokenInfo?.tokenAddress]["usd"],
    }));

    // filter out disallowlist addresses
    const balancesUSDFiltered = balancesUSD.filter(
      (x) => !DISALLOWLISTED_ADDRESSES.includes(x?.tokenAddress)
    );
    const sorted = balancesUSDFiltered.sort((a, b) =>
      a?.tokenBalanceUSD < b?.tokenBalanceUSD ? 1 : -1
    );

    return sorted;
  } catch (e) {
    console.log(e);
  }
}

export async function getTerraCustody(chainInfo, useAllowList) {
  const tokenAccounts = await getTerraTokenAccounts(chainInfo, useAllowList);
  console.log(
    `Num of ${chainInfo?.platform} token accounts=${tokenAccounts?.length}`
  );
  let custody = undefined;
  try {
    custody = await getTokenValues(chainInfo, tokenAccounts, useAllowList);
    console.log(
      `Num of filtered ${chainInfo?.platform} token accounts=${custody?.length}`
    );
  } catch (e) {
    console.log("could not fetch terra prices");
  }
  return custody;
}

export async function grabTerraCustodyData(chain, useAllowList) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  const balances = await getTerraCustody(chainInfo, useAllowList);
  if (balances === undefined) {
    console.log("could not pull terra balances");
    return { balances: [] };
  } else {
    const chainInfo_ = {
      ...chainInfo,
      emitter_address: await getEmitterAddressTerra(
        chainInfo.token_bridge_address
      ),
      balances: balances,
    };
    return chainInfo_;
  }
}

// const chain = process.env.chain;
// const useAllowListstr = process.env.chain || "false";

// var func = async () => {
//   const chainInfo = CHAIN_INFO_MAP[chain];
//   const useAllowList = true ? useAllowListstr === "true" : false;
//   const balances = await getTerraCustody(chainInfo, useAllowList);
//   console.log(balances);
//   await updateTable(chainInfo, balances);
//   // console.log("terra custody (USD) = ", terraCustodyUSD);
// };

// func().then((x) => console.log("end"));
