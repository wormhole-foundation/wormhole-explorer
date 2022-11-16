import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { POLL_TIME } from "../utils/consts";

export type VAAsResponse = {
  indexedAt: string;
  updatedAt: string;
  emitterAddr: string;
  emitterChain: number;
  guardianSetIndex: number;
  sequence: number;
  timestamp: string;
  version: number;
  vaas: string;
  _id: string;
};

function useLatestVAAs(id?: string): VAAsResponse[] {
  const { currentNetwork } = useNetworkContext();
  const [vaas, setVAAs] = useState<VAAsResponse[]>([]);
  useEffect(() => {
    setVAAs([]);
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<VAAsResponse[]>(
          `/api/vaas${id ? `/${id}` : ""}`
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
export default useLatestVAAs;
