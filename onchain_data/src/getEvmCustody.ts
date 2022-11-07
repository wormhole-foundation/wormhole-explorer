import { formatUnits } from "ethers/lib/utils";
import { ethers } from "ethers";
import { getTokenPricesCGID, getTokenPricesGET } from "./getCoinGeckoPrices";

import {
  CHAIN_ID_OASIS,
  isEVMChain,
  CHAIN_ID_KARURA,
  CHAIN_ID_ACALA,
  CHAIN_ID_CELO,
  getEmitterAddressEth,
} from "@certusone/wormhole-sdk";

import { abi as Erc20Abi } from "./abi/erc20.json";

import axios from "axios";
import {
  DISALLOWLISTED_ADDRESSES,
  CHAIN_INFO_MAP,
  newProvider,
  sleepFor,
} from "./utils";

// current allowlist used for stats/govenor
import allowList = require("./allowList.json");

require("dotenv").config();

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

async function getBridgeBalanceScanner(chainInfo) {
  const chainId = chainInfo.chain_id;
  const bridgeAddress = chainInfo.token_bridge_address;
  const url = chainInfo.urlStem;

  // Get native token balances locked in contract
  const balance_apireqstring = `${url}/api?module=account&action=balance&address=${bridgeAddress}`;
  const balance_resp = await axios.get(balance_apireqstring);
  var tokenAccounts = [];
  if (balance_resp.status == 200) {
    const data = balance_resp["data"];
    if (data.hasOwnProperty("result")) {
      const balance = data["result"];
      var nativeInfo = {};
      var symbol = "";
      if (chainId == 7) {
        symbol = "ROSE";
      } else if (chainId == 11) {
        symbol = "KAR";
      } else if (chainId == 12) {
        symbol: "ACA";
      } else if (chainId == 14) {
        symbol: "CELO";
      }
      nativeInfo = {
        decimals: 18,
        name: symbol,
        tokenAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee2",
        balance: balance,
        symbol: symbol,
      };
      tokenAccounts.push(nativeInfo);
    } else {
      console.log("object has no native balance data");
    }
  } else {
    console.log(balance_resp.status);
  }

  // Get locked token list from scan site
  const apiReqString = `${url}/api?module=account&action=tokenlist&address=${bridgeAddress}`;
  console.log(apiReqString);
  const response = await axios.get(apiReqString);
  if (response.status == 200) {
    const data = response["data"];
    if (data.hasOwnProperty("result")) {
      const tokenData = data["result"];
      tokenAccounts = [
        ...tokenAccounts,
        ...tokenData.map((x) => ({
          decimals: x.decimals,
          name: x.name,
          tokenAddress: x.contractAddress,
          balance: x.balance,
          symbol: x.symbol,
        })),
      ];
      return tokenAccounts;
    } else {
      console.log("object has no data");
      return [];
    }
  } else {
    console.log(response.status);
    return [];
  }
}

async function getBridgeBalanceCovalent(chainInfo) {
  /* Get token list (also has bridge balance and prices, but could be stale, so will query onchain)*/

  const covalentChainId: string = chainInfo.covalentChain;
  const bridgeAddress: string = chainInfo.token_bridge_address;
  const covalentApiKey = process.env.REACT_APP_COVALENT_API_KEY;
  const apiReqString = `https://api.covalenthq.com/v1/${covalentChainId}/address/${bridgeAddress}/balances_v2/?quote-currency=USD&format=JSON&nft=true&no-nft-fetch=false&key=${covalentApiKey}`;
  console.log(apiReqString);
  const response = await axios.get(apiReqString, {
    headers: { "User-Agent": "Mozilla/5.0" },
  });
  if (response.status == 200) {
    const data = response["data"];
    if (data.hasOwnProperty("data")) {
      const tokenData = data["data"];
      // tokenData["items"].forEach((x) => console.log(x));
      const tokenAccounts = tokenData["items"]
        .filter((item) => item.type !== "nft")
        .map((x) => ({
          decimals: x.contract_decimals,
          name: x.contract_name,
          tokenAddress: x.contract_address,
          balance: x.balance,
          price: x.quote_rate,
        }));
      return tokenAccounts;
    } else {
      console.log("object has no data");
      return [];
    }
  } else {
    console.log(response.status);
    return [];
  }
}

//for evm chains
async function getTokenContract(
  address: string,
  provider:
    | ethers.providers.JsonRpcProvider
    | ethers.providers.JsonRpcBatchProvider
) {
  const contract = new ethers.Contract(address, Erc20Abi, provider);
  return contract;
}

async function getBridgeBalanceOnChain(chainInfo, tokenList: any[]) {
  const bridgeAddress = chainInfo.token_bridge_address;
  let provider = newProvider(
    chainInfo.endpoint_url,
    true
  ) as ethers.providers.JsonRpcBatchProvider;

  var tokenAccounts = [];
  let i = 0;
  let chunksize = 100;
  while (i < tokenList.length) {
    const tokenContracts = await Promise.all(
      tokenList
        .slice(i, i + chunksize)
        .map((token) => getTokenContract(token.tokenAddress, provider))
    );
    const tokenInfos = await Promise.all(
      tokenContracts.map((tokenContract) =>
        Promise.all([
          tokenContract.address.toLowerCase(),
          tokenContract.name(),
          tokenContract.decimals(),
          tokenContract.symbol(),
          tokenContract.balanceOf(bridgeAddress),
        ])
      )
    );

    tokenInfos.forEach((tokenInfo) => {
      tokenAccounts.push({
        tokenAddress: tokenInfo[0],
        name: tokenInfo[1],
        decimals: tokenInfo[2],
        symbol: tokenInfo[3],
        balance: tokenInfo[4],
      });
    });

    i += chunksize;
  }
  return tokenAccounts;
}

export async function getEvmTokenAccounts(chainInfo, useAllowList) {
  const chainId = chainInfo.chain_id;
  console.log(chainId);
  if (!isEVMChain(chainId)) {
    console.log(`error. ${chainId} is not evm chain`);
    return [];
  }

  var tokenAccounts = [];
  if (useAllowList) {
    const allowList = getAllowList(chainId);
    var tokenList = [];
    Object.keys(allowList).forEach((address) => {
      tokenList.push({
        tokenAddress: address,
      });
    });
    // console.log(tokenList);
    tokenAccounts = await getBridgeBalanceOnChain(
      chainInfo,
      tokenList.filter(
        (token) =>
          token.tokenAddress.toLowerCase() !=
          "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" // filter out native token
      )
    );
    // console.log("tokenAccounts", tokenAccounts);
  } else {
    if (
      chainId == CHAIN_ID_OASIS ||
      chainId == CHAIN_ID_KARURA ||
      chainId == CHAIN_ID_ACALA ||
      chainId == CHAIN_ID_CELO
    ) {
      tokenAccounts = await getBridgeBalanceScanner(chainInfo);
    } else {
      // use convalent to get token addresses
      const tokenList = await getBridgeBalanceCovalent(chainInfo);
      // console.log(tokenList.filter((token) => console.log(token)));
      tokenAccounts = await getBridgeBalanceOnChain(
        chainInfo,
        tokenList.filter(
          (token) =>
            token.tokenAddress.toLowerCase() !=
            "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" // native token
        )
      );
      // console.log(tokenAccounts);
    }
  }
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
      // console.log(tokenPrices);
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

export async function getEvmCustody(chainInfo, useAllowList = true) {
  const tokenAccounts = await getEvmTokenAccounts(chainInfo, useAllowList);
  console.log(
    `Num of ${chainInfo.platform} token accounts=${tokenAccounts.length}`
  );
  const custody = await getTokenValues(chainInfo, tokenAccounts, useAllowList);
  console.log(
    `Num of filtered ${chainInfo.platform} token accounts=${custody.length}`
  );
  return custody;
}

export async function grabEvmCustodyData(chain, useAllowList) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  const balances = await getEvmCustody(chainInfo, useAllowList);
  // await updateTable(chainInfo, balances);
  const chainInfo_ = {
    ...chainInfo,
    emitter_address: getEmitterAddressEth(chainInfo.token_bridge_address),
    balances: balances,
  };
  return chainInfo_;
}

// const chain = process.env.chain;
// const useAllowListstr = process.env.chain || "false";

// (async () => {
//   const chainInfo = CHAIN_INFO_MAP[chain];
//   const useAllowList = true ? useAllowListstr === "true" : false;
//   const balances = await getEvmCustody(chainInfo, useAllowList);
//   console.log(balances);
//   await updateTable(chainInfo, balances);
// })();
