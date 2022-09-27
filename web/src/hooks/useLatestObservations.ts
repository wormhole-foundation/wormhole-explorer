import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { POLL_TIME } from "../utils/consts";

export type ObservationsResponse = {
  createdAt: string;
  updatedAt: string;
  addr: string;
  hash: string;
  messageid: string;
  signature: string;
  txhash: string;
  _id: string;
};

function useLatestObservations(id?: string): ObservationsResponse[] {
  const { currentNetwork } = useNetworkContext();
  const [vaas, setVAAs] = useState<ObservationsResponse[]>([]);
  useEffect(() => {
    setVAAs([]);
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<ObservationsResponse[]>(
          `/api/observations${id ? `/${id}` : ""}?limit=100`
        );
        if (!cancelled) {
          setVAAs(response.data);
          await new Promise((resolve) => setTimeout(resolve, POLL_TIME));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork, id]);
  return vaas;
}
export default useLatestObservations;
