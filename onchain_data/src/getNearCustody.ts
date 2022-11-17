import { formatUnits } from "ethers/lib/utils";
import { BigNumber } from "ethers";
import { getTokenPricesCGID, getTokenPricesGET } from "./getCoinGeckoPrices";

import {
  ChainId,
  CHAIN_ID_NEAR,
  getEmitterAddressNear,
} from "@certusone/wormhole-sdk";

import {
  connect as nearConnect,
  keyStores as nearKeyStores,
  utils as nearUtils,
  Account as nearAccount,
  providers as nearProviders,
} from "near-api-js";

import { CHAIN_INFO_MAP, DISALLOWLISTED_ADDRESSES, sleepFor } from "./utils";

require("dotenv").config();

async function findNearMetadata(userAccount, tokenAddress) {
  if (tokenAddress === "near") {
    //transfering native near
    return {
      decimals: 24,
      name: "NEAR",
      symbol: "NEAR",
    };
  } else {
    const meta_data = await userAccount.viewFunction({
      contractId: tokenAddress,
      methodName: "ft_metadata",
      args: {},
    });
    return meta_data;
  }
}

import allowList = require("./allowList.json");

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
  //   const provider = new nearProviders.JsonRpcProvider(chainInfo.endpoint_url);
  const ACCOUNT_ID = "sender.mainnet";
  var tokenAccounts = [];
  let near = await nearConnect({
    headers: {},
    networkId: "mainnet",
    nodeUrl: chainInfo.endpoint_url,
  });
  tokenAccounts = [];
  const userAccount = new nearAccount(near.connection, bridgeAddress);
  let i = 0;
  while (i < tokenList.length) {
    let tokenAddress = tokenList[i].tokenAddress;
    const tokenInfo = await findNearMetadata(userAccount, tokenAddress);

    let balance = BigNumber.from("0");
    if (tokenAddress === "near") {
      const nearBalanceInfo = await userAccount.getAccountBalance();
      balance = BigNumber.from(nearBalanceInfo.total);
    } else {
      balance = await userAccount.viewFunction({
        contractId: tokenAddress,
        methodName: "ft_balance_of",
        args: {
          account_id: "contract.portalbridge.near",
        },
      });
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

export async function getNearTokenAccounts(chainInfo, useAllowList) {
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

export async function getNearCustody(chainInfo, useAllowList = true) {
  const tokenAccounts = await getNearTokenAccounts(chainInfo, useAllowList);
  console.log(
    `Num of ${chainInfo.platform} token accounts=${tokenAccounts.length}`
  );
  const custody = await getTokenValues(chainInfo, tokenAccounts, useAllowList);
  console.log(
    `Num of filtered ${chainInfo.platform} token accounts=${custody.length}`
  );
  return custody;
}

export async function grabNearCustodyData(chain, useAllowList) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  var balances = [];
  try {
    balances = await getNearCustody(chainInfo, useAllowList);
  } catch (e) {
    console.log("could not grab Near data");
  }
  // await updateTable(chainInfo, balances);
  const chainInfo_ = {
    ...chainInfo,
    emitter_address: getEmitterAddressNear(chainInfo.token_bridge_address),
    balances: balances,
  };
  return chainInfo_;
}

// const chain = process.env.chain;
// const useAllowListstr = process.env.allowlist || "false";

// (async () => {
//   const chainInfo = CHAIN_INFO_MAP[chain];
//   const useAllowList = true ? useAllowListstr === "true" : false;
//   const balances = await getNearCustody(chainInfo, useAllowList);
//   console.log(balances);
// })();
