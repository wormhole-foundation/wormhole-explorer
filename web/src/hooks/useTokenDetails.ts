import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { POLL_TIME } from "../utils/consts";

export type TokenDetailsResponse = {
  tokenAddress: string;
  name: string;
  decimals: number;
  symbol: string;
  balance: BigInt;
  qty: number;
  tokenPrice: number;
  tokenBalanceUSD: number;
};

export type TokensResponse = {
  _id: string;
  tokens: TokenDetailsResponse[];
};

function useTokenDetails(id?: string): TokenDetailsResponse[] {
  const { currentNetwork } = useNetworkContext();
  const [tokenDetails, setTokenDetails] = useState<TokenDetailsResponse[]>([]);
  useEffect(() => {
    setTokenDetails([]);
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<TokensResponse>(
          `/api/custody/tokens${id ? `/${id}` : ""}`
        );
        if (!cancelled) {
          setTokenDetails(response.data.tokens);
          await new Promise((resolve) => setTimeout(resolve, POLL_TIME));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork, id]);
  return tokenDetails;
}
export default useTokenDetails;
