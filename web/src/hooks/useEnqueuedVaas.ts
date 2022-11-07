import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { POLL_TIME } from "../utils/consts";

export type EnqueuedVaa = {
  chainId: number;
  emitterAddres: string;
  sequence: number;
  notionalValue: number;
  txHash: string;
};

export type EnqueuedVaasResponse = {
  chainId: string;
  enqueuedVaas: EnqueuedVaa[];
};

function useEnqueuedVaas(id?: string): EnqueuedVaasResponse[] {
  const { currentNetwork } = useNetworkContext();
  const [enqueuedVaas, setEnqueuedVaas] = useState<EnqueuedVaasResponse[]>([]);
  useEffect(() => {
    setEnqueuedVaas([]);
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<EnqueuedVaasResponse[]>(
          `/api/enqueuedVaas`
        );
        if (!cancelled) {
          setEnqueuedVaas(response.data);
          await new Promise((resolve) => setTimeout(resolve, POLL_TIME));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork, id]);
  return enqueuedVaas;
}
export default useEnqueuedVaas;
