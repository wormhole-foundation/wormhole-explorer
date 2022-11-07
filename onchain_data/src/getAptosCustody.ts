import { formatUnits } from "ethers/lib/utils";
import { BigNumber } from "ethers";
import { getTokenPricesCGID, getTokenPricesGET } from "./getCoinGeckoPrices";
import allowList = require("./allowList.json");
import { AptosClient } from "aptos";

import {
  getTypeFromExternalAddress,
  isValidAptosType,
} from "@certusone/wormhole-sdk";

import { CHAIN_INFO_MAP, DISALLOWLISTED_ADDRESSES } from "./utils";

require("dotenv").config();

async function findAptosMetadata(provider, tokenAddress) {
  let decimals = 0;
  let name = undefined;
  let symbol = undefined;

  try {
    const coinInfo = await provider.getAccountResource(
      tokenAddress.split("::")[0],
      `0x1::coin::CoinInfo<${tokenAddress}>`
    );
    const metaData = coinInfo.data;
    decimals = metaData["decimals"];
    name = metaData["name"];
    symbol = metaData["symbol"];
  } catch (e) {
    console.log(`could not find meta_data for address=${tokenAddress}`);
  }

  return {
    tokenAddress: tokenAddress,
    decimals: Number(decimals),
    name: name,
    symbol: symbol,
  };
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

async function getBridgeBalanceOnChain(chainInfo, tokenList: any[]) {
  const bridgeAddress = chainInfo.token_bridge_address;
  var tokenAccounts = [];
  const provider = new AptosClient(chainInfo.endpoint_url);
  tokenAccounts = [];
  let i = 0;
  while (i < tokenList.length) {
    let tokenAddress = tokenList[i].tokenAddress;
    let tokenAddressQualified = tokenAddress;
    if (!isValidAptosType(tokenAddress)) {
      tokenAddressQualified = await getTypeFromExternalAddress(
        provider,
        bridgeAddress,
        tokenAddress
      );
      console.log("converting hash to aptos type", tokenAddressQualified);
    }
    const tokenInfo = await findAptosMetadata(provider, tokenAddressQualified);
    let balance = BigNumber.from("0");
    try {
      const accountResource = await provider.getAccountResource(
        bridgeAddress,
        `0x1::coin::CoinStore<${tokenAddressQualified}>`
      );
      balance = accountResource?.data["coin"]["value"];
    } catch {
      console.error(`could not get aptos balance for ${tokenAddressQualified}`);
    }
    tokenAccounts.push({
      tokenAddress: tokenAddress,
      name: tokenInfo.name,
      decimals: tokenInfo.decimals,
      symbol: tokenInfo.symbol,
      balance: BigNumber.from(balance),
    });
    i++;
  }
  return tokenAccounts;
}

export async function getAptosTokenAccounts(chainInfo, useAllowList) {
  const chainId = chainInfo.chain_id;
  console.log(chainId);

  var tokenAccounts = [];
  if (useAllowList) {
    const allowList = getAllowList(chainId);
    var tokenList = [];
    Object.keys(allowList).forEach((address) => {
      tokenList.push({
        tokenAddress: address,
      });
    });
    tokenAccounts = await getBridgeBalanceOnChain(chainInfo, tokenList);
  } /*else {
            // pull from transactions table
        tokenAccounts = await getBridgeBalanceScanner(chainInfo);
    }*/
  return tokenAccounts;
}

async function getTokenValues(chainInfo, tokenInfos: any[], useAllowList) {
  console.log("allowlist?", useAllowList);
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
      // use tokenAddresses from tokenInfos/custody

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
      Object.keys(tokenPrices).includes(x.tokenAddress)
    );
    // calculate usd balances. add price and usd balance to tokenInfos
    const balancesUSD = filteredBalances.map((tokenInfo) => ({
      ...tokenInfo,
      tokenPrice: tokenPrices[tokenInfo.tokenAddress]["usd"],
      tokenBalanceUSD:
        tokenInfo.qty * tokenPrices[tokenInfo.tokenAddress]["usd"],
    }));

    // filter out disallowlist addresses
    const balancesUSDFiltered = balancesUSD.filter(
      (x) => !DISALLOWLISTED_ADDRESSES.includes(x.tokenAddress)
    );
    const sorted = balancesUSDFiltered.sort((a, b) =>
      a.tokenBalanceUSD < b.tokenBalanceUSD ? 1 : -1
    );

    return sorted;
  } catch (e) {
    console.log(e);
  }
  return [];
}

export async function getAptosCustody(chainInfo, useAllowList = true) {
  const tokenAccounts = await getAptosTokenAccounts(chainInfo, useAllowList);
  console.log(
    `Num of ${chainInfo.platform} token accounts=${tokenAccounts.length}`
  );
  const custody = await getTokenValues(chainInfo, tokenAccounts, useAllowList);
  console.log(
    `Num of filtered ${chainInfo.platform} token accounts=${custody.length}`
  );
  return custody;
}

export async function grabAptosCustodyData(chain, useAllowList) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  const balances = await getAptosCustody(chainInfo, useAllowList);
  // await updateTable(chainInfo, balances);
  const chainInfo_ = {
    ...chainInfo,
    emitter_address:
      "0000000000000000000000000000000000000000000000000000000000000001",
    balances: balances,
  };
  return chainInfo_;
}

// const chain = process.env.chain;
// const useAllowListstr = process.env.allowlist || "false";

// (async () => {
//   const chainInfo = CHAIN_INFO_MAP[chain];
//   const useAllowList = true ? useAllowListstr === "true" : false;
//   const balances = await getAptosCustody(chainInfo, useAllowList);
//   console.log(balances);
// })();
