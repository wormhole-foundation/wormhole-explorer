import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { POLL_TIME } from "../utils/consts";

export type Token = {
  tokenAddress: string;
  name: string;
  decimals: number;
  symbol: string;
  balance: BigInt;
  qty: number;
  tokenPrice: number;
  tokenBalanceUSD: number;
};

export type CustodyDataResponse = {
  _id: string;
  updatedAt: string;
  chainId: number;
  chainName: string;
  custodyUSD: number;
  emitterAddress: string;
  tokens: Token[];
};

function useCustodyData(): CustodyDataResponse[] {
  const { currentNetwork } = useNetworkContext();
  const [custodyData, setCustodyData] = useState<CustodyDataResponse[]>([]);
  useEffect(() => {
    setCustodyData([]);
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<CustodyDataResponse[]>(`/api/custody`);
        if (!cancelled) {
          setCustodyData(response.data);
          await new Promise((resolve) => setTimeout(resolve, POLL_TIME));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork]);
  return custodyData;
}
export default useCustodyData;
