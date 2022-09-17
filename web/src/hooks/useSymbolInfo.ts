import {
  ChainId,
  CHAIN_ID_SOLANA,
  CHAIN_ID_TERRA,
  isEVMChain,
  isHexNativeTerra,
  tryHexToNativeAssetString,
  tryNativeToHexString,
} from "@certusone/wormhole-sdk";
import { GovernorGetTokenListResponse_Entry } from "@certusone/wormhole-sdk-proto-web/lib/cjs/publicrpc/v1/publicrpc";
import { LCDClient } from "@terra-money/terra.js";
import axios from "axios";
import { ethers } from "ethers";
import { useEffect, useMemo } from "react";
import { CHAIN_INFO_MAP } from "../utils/consts";

require("dotenv").config();

function newProvider(
  url: string,
  batch: boolean = false
): ethers.providers.JsonRpcProvider | ethers.providers.JsonRpcBatchProvider {
  // only support http(s), not ws(s) as the websocket constructor can blow up the entire process
  // it uses a nasty setTimeout(()=>{},0) so we are unable to cleanly catch its errors
  if (url.startsWith("http")) {
    if (batch) {
      return new ethers.providers.JsonRpcBatchProvider(url);
    }
    return new ethers.providers.JsonRpcProvider(url);
  }
  throw new Error("url does not start with http/https!");
}

const ERC20_BASIC_ABI = [
  "function name() view returns (string name)",
  "function symbol() view returns (string symbol)",
  "function decimals() view returns (uint8 decimals)",
];

//for evm chains
async function getTokenContract(
  address: string,
  provider:
    | ethers.providers.JsonRpcProvider
    | ethers.providers.JsonRpcBatchProvider
) {
  const contract = new ethers.Contract(address, ERC20_BASIC_ABI, provider);
  return contract;
}

type TokenInfo = {
  address: string;
  name: string;
  decimals: number;
  symbol: string;
};

async function getEvmTokenMetaData(
  chain: ChainId,
  governorTokenList: GovernorGetTokenListResponse_Entry[]
) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  var formattedAddressMap: { [key: string]: string } = {};
  var formattedTokens: string[] = [];
  try {
    const tokenListByChain = governorTokenList
      .filter(
        (g) =>
          g.originChainId === chainInfo.chainId &&
          !TOKEN_CACHE.get([chain, g.originAddress].join("_"))
      )
      .map((tk) => tk.originAddress);
    tokenListByChain.forEach((tk) => {
      const tk_ = tryHexToNativeAssetString(tk.slice(2), chainInfo.chainId);
      formattedTokens.push(tk_);
      formattedAddressMap[tk_] = tk;
    });

    let provider = newProvider(
      chainInfo.endpointUrl,
      true
    ) as ethers.providers.JsonRpcBatchProvider;

    const tokenContracts = await Promise.all(
      formattedTokens.map((tokenAddress) =>
        getTokenContract(tokenAddress, provider)
      )
    );
    const tokenInfos = await Promise.all(
      tokenContracts.map((tokenContract) =>
        Promise.all([
          tokenContract.address.toLowerCase(),
          tokenContract.name(),
          tokenContract.decimals(),
          // tokenContract.balanceOf(bridgeAddress),
          tokenContract.symbol(),
        ])
      )
    );

    tokenInfos.forEach((tk) => {
      TOKEN_CACHE.set([chain, formattedAddressMap[tk[0]]].join("_"), {
        address: tk[0],
        name: tk[1],
        decimals: tk[2],
        symbol: tk[3],
      });
    });
  } catch (e) {
    console.log(e);
    console.log(chain, chainInfo);
  }
  return;
}

async function loadTokenListSolana(
  url = "https://token-list.solana.com/solana.tokenlist.json"
) {
  const response: any = await axios.get(url).catch(function (error) {
    if (error.response) {
      console.log("could not load token list", error.response.status);
    }
  });
  if (response["status"] === 200) {
    const data = response["data"];
    const token_list = data.tokens;
    return token_list;
  } else {
    console.log("bad response for token list");
    return [];
  }
}

async function getSolanaTokenMetaData(
  chain: ChainId,
  governorTokenList: GovernorGetTokenListResponse_Entry[]
) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  var formattedAddressMap: { [key: string]: string } = {};
  var formattedTokens: string[] = [];
  try {
    const tokenListByChain = governorTokenList
      .filter(
        (g) =>
          g.originChainId === chainInfo.chainId &&
          !TOKEN_CACHE.get([chain, g.originAddress].join("_"))
      )
      .map((tk) => tk.originAddress);

    tokenListByChain.forEach((tk) => {
      const tk_ = tryHexToNativeAssetString(tk.slice(2), chainInfo.chainId);
      formattedTokens.push(tk_);
      formattedAddressMap[tk_] = tk;
    });

    var metaDataArr: any[] = [];
    try {
      metaDataArr = await loadTokenListSolana();
    } catch (e) {
      console.log(e);
    }
    var tokenContracts: TokenInfo[] = [];
    formattedTokens.forEach((token) => {
      const metaData = metaDataArr.filter((x) => x.address === token);
      if (metaData.length > 0) {
        tokenContracts.push(metaData[0]);
      }
    });
    tokenContracts.forEach((tokenContract) => {
      TOKEN_CACHE.set(
        [chain, formattedAddressMap[tokenContract.address]].join("_"),
        {
          address: tokenContract.address,
          name: tokenContract.name,
          decimals: tokenContract.decimals,
          symbol: tokenContract.symbol,
        }
      );
    });
  } catch (e) {
    console.log(e);
    console.log(chain, chainInfo);
  }
  return;
}

type TerraMetadata = {
  address: string;
  symbol?: string;
  logo?: string;
  name?: string;
  decimals?: number;
};

const fetchSingleTerraMetadata = async (address: string, lcd: LCDClient) =>
  lcd.wasm
    .contractQuery(address, {
      token_info: {},
    })
    .then(
      ({ symbol, name, decimals }: any) =>
        ({ address: address, symbol, name, decimals } as TerraMetadata)
    );

async function getSingleTerraMetaData(
  originAddress: string,
  originChain: ChainId
) {
  const TERRA_HOST = {
    URL:
      originChain === CHAIN_ID_TERRA
        ? "https://columbus-fcd.terra.dev"
        : "https://phoenix-fcd.terra.dev",
    chainID: originChain === CHAIN_ID_TERRA ? "columbus-5" : "phoenix-1",
    name: "mainnet",
  };
  const lcd = new LCDClient(TERRA_HOST);

  if (isHexNativeTerra(tryNativeToHexString(originAddress, originChain))) {
    if (originAddress === "uusd") {
      return {
        address: originAddress,
        name: "UST Classic",
        symbol: "USTC",
        decimals: 6,
      };
    } else if (originAddress === "uluna") {
      return {
        address: originAddress,
        name: "Luna Classic",
        symbol: "LUNC",
        decimals: 6,
      };
    } else {
      return {
        address: originAddress,
        name: "",
        symbol: "",
        decimals: 8,
      };
    }
  } else {
    return await fetchSingleTerraMetadata(originAddress, lcd);
  }
}

async function getTerraTokenMetaData(
  chain: ChainId,
  governorTokenList: GovernorGetTokenListResponse_Entry[]
) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  var formattedAddressMap: { [key: string]: string } = {};
  var formattedTokens: string[] = [];
  try {
    const tokenListByChain = governorTokenList
      .filter(
        (g) =>
          g.originChainId === chainInfo.chainId &&
          !TOKEN_CACHE.get([chain, g.originAddress].join("_"))
      )
      .map((tk) => tk.originAddress);

    tokenListByChain.forEach((tk) => {
      const tk_ = tryHexToNativeAssetString(tk.slice(2), chainInfo.chainId);
      formattedTokens.push(tk_);
      formattedAddressMap[tk_] = tk;
    });

    var tokenContracts: any[] = [];
    for (let i = 0; i < formattedTokens.length; i++) {
      const token = formattedTokens[i];
      const metaData = await getSingleTerraMetaData(token, chain);
      tokenContracts.push(metaData);
    }

    tokenContracts.forEach((tokenContract) => {
      TOKEN_CACHE.set(
        [chain, formattedAddressMap[tokenContract.address]].join("_"),
        {
          address: tokenContract.address,
          name: tokenContract.name,
          decimals: tokenContract.decimals,
          symbol: tokenContract.symbol,
        }
      );
    });
  } catch (e) {
    console.log(e);
    console.log(chain, chainInfo);
  }
  return;
}

const MISC_TOKEN_META_DATA: {
  [key: string]: {
    [key: string]: { name: string; symbol: string; decimals: number };
  };
} = {
  "8": {
    "0x0000000000000000000000000000000000000000000000000000000000000000": {
      name: "ALGO",
      symbol: "ALGO",
      decimals: 6,
    },
    "0x000000000000000000000000000000000000000000000000000000000004c5c1": {
      name: "USDT",
      symbol: "USDT",
      decimals: 6,
    },
    "0x0000000000000000000000000000000000000000000000000000000001e1ab70": {
      name: "USDC",
      symbol: "USDC",
      decimals: 6,
    },
  },
  "15": {
    "0x0000000000000000000000000000000000000000000000000000000000000000": {
      name: "NEAR",
      symbol: "NEAR",
      decimals: 24,
    },
  },
  "18": {
    "0x01fa6c6fbc36d8c245b0a852a43eb5d644e8b4c477b27bfab9537c10945939da": {
      name: "LUNA",
      symbol: "LUNA",
      decimals: 6,
    },
  },
};

async function getMiscTokenMetaData(
  chain: ChainId,
  governorTokenList: GovernorGetTokenListResponse_Entry[]
) {
  const chainInfo = CHAIN_INFO_MAP[chain];
  const tokenMetaDataByChain = MISC_TOKEN_META_DATA[chain.toString()];
  try {
    const tokenListByChain = governorTokenList
      .filter(
        (g) =>
          g.originChainId === chainInfo.chainId &&
          !TOKEN_CACHE.get([chain, g.originAddress].join("_"))
      )
      .map((tk) => tk.originAddress);

    tokenListByChain.forEach((tk) => {
      const metaData = tokenMetaDataByChain[tk];
      TOKEN_CACHE.set([chain, tk].join("_"), {
        address: tk,
        name: metaData?.name,
        decimals: metaData?.decimals,
        symbol: metaData?.symbol,
      });
    });
  } catch (e) {
    console.log(e);
    console.log(chain, chainInfo);
  }
  return;
}

const TOKEN_CACHE = new Map<string, TokenInfo>();

async function getTokenMetaData(
  governorTokenList: GovernorGetTokenListResponse_Entry[]
) {
  const chains = Object.keys(CHAIN_INFO_MAP);
  for (let i = 0; i < chains.length; i++) {
    const chain = chains[i];

    const chainInfo = CHAIN_INFO_MAP[chain];
    const chainId = chainInfo.chainId;
    try {
      //grab token info

      if (isEVMChain(chainId)) {
        await getEvmTokenMetaData(chainId, governorTokenList);
      } else if (chainId === CHAIN_ID_SOLANA) {
        await getSolanaTokenMetaData(chainId, governorTokenList);
      } else if (chainId === CHAIN_ID_TERRA) {
        await getTerraTokenMetaData(chainId, governorTokenList);
      } else {
        // currently no support for ALGORAND, NEAR, TERRA2
        console.log(`the chain=${chain} is not supported`);
        await getMiscTokenMetaData(chainId, governorTokenList);
      }
      await new Promise((resolve) => setTimeout(resolve, 6000));
    } catch (e) {
      console.log(e);
      console.log(chain, chainInfo);
      continue;
    }
  }

  await new Promise((resolve) => setTimeout(resolve, 3000000));
  return;
}

const TIMEOUT = 60 * 1000;

function useSymbolInfo(tokens: GovernorGetTokenListResponse_Entry[]) {
  // TODO: GovernorInfo gets fetched repeatedly, but we don't need to refresh the list
  // So using string'd version of token list as a hack around when to update the token list
  const memoizedTokens = useMemo(() => JSON.stringify(tokens), [tokens]);

  useEffect(() => {
    const tokens = JSON.parse(memoizedTokens);
    (async () => {
      // TODO: use a state setter to update instead of relying on TOKEN_CACHE.
      await getTokenMetaData(tokens);
      await new Promise((resolve) => setTimeout(resolve, TIMEOUT));
    })();
  }, [memoizedTokens]);
  return TOKEN_CACHE;
}

export default useSymbolInfo;
