import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { POLL_TIME } from "../utils/consts";

export type GovernorStatusChainDetailsResponse = {
  _id: string;
  createdAt: string;
  updatedAt: number;
  nodeName: string;
  chainId: number;
  availableNotional: number;
};

function useGovernorStatusChainDetails(
  id?: string
): GovernorStatusChainDetailsResponse[] {
  const { currentNetwork } = useNetworkContext();
  const [governorStatusChainDetails, setGovernorStatusChainDetails] = useState<
    GovernorStatusChainDetailsResponse[]
  >([]);
  useEffect(() => {
    setGovernorStatusChainDetails([]);
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<GovernorStatusChainDetailsResponse[]>(
          `/api/availableNotional${id ? `/${id}` : ""}`
        );
        if (!cancelled) {
          setGovernorStatusChainDetails(response.data);
          await new Promise((resolve) => setTimeout(resolve, POLL_TIME));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork, id]);
  return governorStatusChainDetails;
}
export default useGovernorStatusChainDetails;
