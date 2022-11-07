import { formatUnits } from "ethers/lib/utils";
import { TOKEN_PROGRAM_ID } from "@solana/spl-token";
import { Connection, AccountInfo, PublicKey } from "@solana/web3.js";
import { Connection as ConnectionMeta, programs } from "@metaplex/js";
import { getTokenPricesCGID, getTokenPricesGET } from "./getCoinGeckoPrices";

import { CHAIN_INFO_MAP, DISALLOWLISTED_ADDRESSES, sleepFor } from "./utils";

// current allowlist used for stats/govenor
import allowList = require("./allowList.json");
import { getEmitterAddressSolana } from "@certusone/wormhole-sdk";
import { BigNumber } from "ethers";

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

const {
  metadata: { Metadata },
} = programs;

export async function getMultipleAccountsRPC(
  connection: Connection,
  pubkeys: PublicKey[]
): Promise<(AccountInfo<Buffer> | null)[]> {
  return getMultipleAccounts(connection, pubkeys, "confirmed");
}

export const getMultipleAccounts = async (
  connection: any,
  pubkeys: PublicKey[],
  commitment: string
) => {
  return (
    await Promise.all(connection.getMultipleAccountsInfo(pubkeys, commitment))
  ).flat();
};

export function shortenAddress(address: string) {
  return address.length > 10
    ? `${address.slice(0, 4)}...${address.slice(-4)}`
    : address;
}

export const METADATA_REPLACE = new RegExp("\u0000", "g");
export const EDITION_MARKER_BIT_SIZE = 248;
export const METADATA_PREFIX = "metadata";
export const EDITION = "edition";

export function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export type StringPublicKey = string;

export enum MetadataKey {
  Uninitialized = 0,
  MetadataV1 = 4,
  EditionV1 = 1,
  MasterEditionV1 = 2,
  MasterEditionV2 = 6,
  EditionMarker = 7,
}

export const METADATA_PROGRAM_ID =
  "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s" as StringPublicKey;

export const getMetadataAddress = async (
  mintKey: string
): Promise<[PublicKey, number]> => {
  const seeds = [
    Buffer.from("metadata"),
    new PublicKey(METADATA_PROGRAM_ID).toBuffer(),
    new PublicKey(mintKey).toBuffer(),
  ];
  return PublicKey.findProgramAddress(
    seeds,
    new PublicKey(METADATA_PROGRAM_ID)
  );
};

async function getSolanaMetaData(
  mintAddresses: string[],
  connection: Connection
) {
  const metaAddresses = [];
  for (let i = 0; i < mintAddresses.length; i++) {
    let mintAddress = mintAddresses[i];
    try {
      const metaAddress = await getMetadataAddress(mintAddress);
      metaAddresses.push({
        mint: mintAddress,
        meta: metaAddress[0].toString(),
      });
    } catch (e) {
      continue;
    }
  }

  let storeMetadata = {};
  // Get store metadata
  for (let i = 0; i < metaAddresses.length; i++) {
    let mintkey = new PublicKey(metaAddresses[i].mint);
    let pubkey = new PublicKey(metaAddresses[i].meta);
    try {
      const metadata = await Metadata.load(connection, pubkey);
      const metadatadata = metadata.data?.data || {};
      metadatadata["metakey"] = pubkey.toString();
      metadatadata["tokenAddress"] = mintkey.toString();
      storeMetadata[mintkey.toString()] = metadatadata;
    } catch (e) {
      continue;
    }
  }
  return storeMetadata;
}

export async function getSolanaTokenAccounts(chainInfo, useAllowList) {
  const chainId = chainInfo.chain_id;
  const custodyAddress = chainInfo.custody_address;
  const connection = new Connection(chainInfo.endpoint_url);

  try {
    var mintAddresses = [];
    var tokenAccounts = [];

    if (useAllowList) {
      const allowList = getAllowList(chainId);

      var mintAddresses = [];
      Object.keys(allowList).forEach((address) => {
        mintAddresses.push(address);
      });
      for (let i = 0; i < mintAddresses.length; i++) {
        const mintAddress = mintAddresses[i];
        const parsedAccount = await connection.getParsedTokenAccountsByOwner(
          new PublicKey(custodyAddress),
          {
            mint: new PublicKey(mintAddress),
          }
        );

        const tokenAccount_ = parsedAccount.value.at(-1);
        const tokenAccount = tokenAccount_.account.data.parsed.info;
        if (tokenAccount.tokenAmount?.amount > 0) {
          // console.log(
          //   `token=${mintAddress} has a balance=${tokenAccount.tokenAmount?.amount}`
          // );
          tokenAccounts.push(tokenAccount);
        } else {
          console.log(`${tokenAccount} has a 0 balance`);
        }
        await sleepFor(1000);
      }
    } else {
      const allAccounts = await connection.getParsedTokenAccountsByOwner(
        new PublicKey(custodyAddress),
        { programId: TOKEN_PROGRAM_ID },
        "confirmed"
      );
      allAccounts.value.forEach((account) => {
        if (account.account.data.parsed?.info?.tokenAmount?.amount > 0) {
          // get all token accounts with nonzero balance
          tokenAccounts.push(account.account.data.parsed?.info);
        }
      });

      // get mint addresses from token accounts and find metadata address
      mintAddresses = tokenAccounts.map((x) => x.mint);
    }

    const storeMetadata = await getSolanaMetaData(mintAddresses, connection);
    tokenAccounts = tokenAccounts.map((custody) => ({
      ...custody,
      ...custody["tokenAmount"],
      ...storeMetadata[custody.mint],
    }));

    // do a little tidying up
    tokenAccounts = tokenAccounts.map((account) => ({
      tokenAddress: account["tokenAddress"],
      name: account["name"],
      decimals: account["decimals"],
      symbol: account["symbol"],
      balance: BigNumber.from(account["amount"]),
    }));
    return tokenAccounts;
  } catch (e) {
    console.log(e);
  }

  return [];
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
      // console.log(prices);
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
        const prices = await getTokenPricesGET(
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
      // console.log("tokenPrices", tokenPrices);
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
    var sorted = balancesUSDFiltered.sort((a, b) =>
      a.tokenBalanceUSD < b.tokenBalanceUSD ? 1 : -1
    );

    return sorted;
  } catch (e) {
    console.log(e);
  }
  return [];
}

export async function getSolanaCustody(chainInfo, useAllowList = true) {
  const endpoint = chainInfo.endpoint_url;
  const bridgeAddress = chainInfo.token_bridge_address;
  const connection = new Connection(endpoint);
  const nativeBalance = await connection.getBalance(
    new PublicKey(bridgeAddress)
  );
  // console.log("amount of sol in custody=", nativeBalance);
  const solanaCustodyNativeUSD = 0; //solanaCustodyNative.map(x => (parseInt(x.balance) / (10.0 ** x.decimals) * x.price)).reduce((partialSum, a) => partialSum + a, 0);

  // grab token accounts from rpc/allowlist
  const tokenAccounts = await getSolanaTokenAccounts(chainInfo, useAllowList);
  console.log(
    `Num of ${chainInfo.platform} token accounts=${tokenAccounts.length}`
  );
  // tokenAccounts.forEach((x) => console.log(x));
  const custody = await getTokenValues(chainInfo, tokenAccounts, useAllowList);
  console.log(
    `Num of filtered ${chainInfo.platform} token accounts=${custody.length}`
  );
  return custody;
}

export async function grabSolanaCustodyData(chain, useAllowList) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  const balances = await getSolanaCustody(chainInfo, useAllowList);
  if (balances.length === 0) {
    console.log(`could not get ${chainInfo.name} custody data`);
  }
  const chainInfo_ = {
    ...chainInfo,
    emitter_address:
      "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
    balances: balances,
  };
  return chainInfo_;
}

// const useAllowListstr = process.env.chain || "false";

// (async () => {
//   const useAllowList = true ? useAllowListstr === "true" : false;
//   const chainInfo = CHAIN_INFO_MAP["1"];
//   const balances = await getSolanaCustody(chainInfo, useAllowList);
//   await updateTable(chainInfo, balances);
// })();
