import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { POLL_TIME } from "../utils/consts";

export type EnqueuedVaaDetailsResponse = {
  chainId: string;
  emitterAddress: string;
  sequence: number;
  notionalValue: number;
  txHash: string;
  releaseTime: number;
};

function useEnqueuedVaaDetails(id?: string): EnqueuedVaaDetailsResponse[] {
  const { currentNetwork } = useNetworkContext();
  const [enqueuedVaaDetails, setEnqueuedVaaDetails] = useState<
    EnqueuedVaaDetailsResponse[]
  >([]);
  useEffect(() => {
    setEnqueuedVaaDetails([]);
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<EnqueuedVaaDetailsResponse[]>(
          `/api/enqueuedVaas${id ? `/${id}` : ""}`
        );
        if (!cancelled) {
          setEnqueuedVaaDetails(response.data);
          await new Promise((resolve) => setTimeout(resolve, POLL_TIME));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork, id]);
  return enqueuedVaaDetails;
}
export default useEnqueuedVaaDetails;
