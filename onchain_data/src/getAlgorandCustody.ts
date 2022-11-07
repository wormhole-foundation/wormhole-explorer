import { formatUnits } from "ethers/lib/utils";
import { BigNumber } from "ethers";
import { getTokenPricesCGID, getTokenPricesGET } from "./getCoinGeckoPrices";

import { getEmitterAddressAlgorand } from "@certusone/wormhole-sdk";
import { calcLogicSigAccount } from "@certusone/wormhole-sdk/lib/cjs/algorand";

export const ALGORAND_HOST = {
  algodToken: "",
  algodServer: "https://mainnet-api.algonode.cloud", //"https://mainnet-idx.algonode.cloud", //"https://mainnet-api.algonode.cloud",
  algodPort: "",
};

import { CHAIN_INFO_MAP, DISALLOWLISTED_ADDRESSES, sleepFor } from "./utils";

require("dotenv").config();

import allowList = require("./allowList.json");
import algosdk from "algosdk";

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

async function findAlgoMetadata(provider, tokenAddress) {
  const index = parseInt(tokenAddress);
  if (index === 0) {
    const definition = {
      tokenAddress: tokenAddress,
      decimals: Number(6),
      name: "ALGO",
      symbol: "ALGO",
    };
    return definition;
  }

  let decimals = undefined;
  let name = undefined;
  let symbol = undefined;

  try {
    const resp = await provider.getAssetByID(index).do();
    decimals = resp.params.decimals;
    name = resp.params.name;
    symbol = resp.params["unit-name"];
  } catch {
    console.error(`could not get ${index} metadata`);
  }
  return {
    tokenAddress: tokenAddress,
    decimals: decimals,
    name: name,
    symbol: symbol,
  };
}

export async function getNativeAlgoAddress(
  algoClient: any,
  token_bridge: any,
  assetId: any
) {
  const { doesExist, lsa } = await calcLogicSigAccount(
    algoClient,
    BigInt(token_bridge),
    BigInt(assetId),
    Buffer.from("native", "binary").toString("hex")
  );
  return lsa.address();
}

async function getBridgeBalanceOnChain(chainInfo, tokenList: any[]) {
  const bridgeAddress = chainInfo.token_bridge_address;
  var tokenAccounts = [];
  const provider = new algosdk.Algodv2(
    ALGORAND_HOST.algodToken,
    ALGORAND_HOST.algodServer,
    ALGORAND_HOST.algodPort
  );
  tokenAccounts = [];
  let i = 0;
  while (i < tokenList.length) {
    let tokenAddress = tokenList[i].tokenAddress;
    const tokenInfo = await findAlgoMetadata(provider, tokenAddress);

    let balance = BigNumber.from("0");
    try {
      let nativeAlgoAddr = await getNativeAlgoAddress(
        provider,
        bridgeAddress,
        parseInt(tokenAddress)
      );

      let nativeAlgoInfo = await provider
        .accountInformation(nativeAlgoAddr)
        .do();
      if (tokenAddress == 0) {
        // native ALGO balance
        balance = nativeAlgoInfo?.amount;
      } else {
        // other native asset
        if (nativeAlgoInfo?.assets.length == 0) {
          balance = BigNumber.from("0");
        } else {
          balance = nativeAlgoInfo?.assets[0].amount;
        }
      }
    } catch {
      console.error(`could not get ${tokenAddress} balance`);
    }
    tokenAccounts.push({
      tokenAddress: tokenAddress,
      name: tokenInfo.name,
      decimals: tokenInfo.decimals,
      symbol: tokenInfo.symbol,
      balance: balance,
    });
    i++;
  }
  return tokenAccounts;
}

export async function getAlgoTokenAccounts(chainInfo, useAllowList) {
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
    // console.log("tokenAccounts", tokenAccounts);
  } /*else {
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

export async function getAlgoCustody(chainInfo, useAllowList = true) {
  const tokenAccounts = await getAlgoTokenAccounts(chainInfo, useAllowList);
  console.log(
    `Num of ${chainInfo.platform} token accounts=${tokenAccounts.length}`
  );
  const custody = await getTokenValues(chainInfo, tokenAccounts, useAllowList);
  console.log(
    `Num of filtered ${chainInfo.platform} token accounts=${custody.length}`
  );
  return custody;
}

export async function grabAlgoCustodyData(chain, useAllowList) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  const balances = await getAlgoCustody(chainInfo, useAllowList);
  // await updateTable(chainInfo, balances);
  const chainInfo_ = {
    ...chainInfo,
    emitter_address: getEmitterAddressAlgorand(
      BigInt(chainInfo.token_bridge_address)
    ),
    balances: balances,
  };
  return chainInfo_;
}

// const chain = process.env.chain;
// const useAllowListstr = process.env.allowlist || "false";

// (async () => {
//   const chainInfo = CHAIN_INFO_MAP[chain];
//   const useAllowList = true ? useAllowListstr === "true" : false;
//   const balances = await getAlgoCustody(chainInfo, useAllowList);
//   console.log(balances);
// })();
