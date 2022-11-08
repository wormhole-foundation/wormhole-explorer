import axios from "axios";
import { useEffect, useState } from "react";
import { POLL_TIME } from "../utils/consts";

export type EnqueuedVaa = {
  sequence: number;
  releasetime: number;
  notionalvalue: number;
  txhash: string;
};

export type Emitter = {
  emitteraddress: string;
  totalequeuedvaas: number;
  enqueuedvaas: EnqueuedVaa[] | null;
};

export type Chain = {
  chainId: number;
  remainingavailablenotional: number;
  emitters: Emitter[];
};

export type GovernorStatusResponse = {
  chainId: number;
  availableNotional: number;
  notionalLimit: number;
  maxTransactionSize: number;
  // emitters: Emitter[];
};

function useGovernorStatus(): GovernorStatusResponse[] {
  // const { currentNetwork } = useNetworkContext();
  const [governorStatus, setGovernorStatus] = useState<
    GovernorStatusResponse[]
  >([]);
  useEffect(() => {
    setGovernorStatus([]);
  }, []);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<GovernorStatusResponse[]>(
          "/api/governorLimits"
        );
        if (!cancelled) {
          console.log(response.data);
          setGovernorStatus(response.data);
          await new Promise((resolve) => setTimeout(resolve, POLL_TIME));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);
  return governorStatus;
}
export default useGovernorStatus;
