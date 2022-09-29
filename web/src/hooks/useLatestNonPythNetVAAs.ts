import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { POLL_TIME } from "../utils/consts";
import { VAAsResponse } from "./useLatestVAAs";

function useLatestNonPythNetVAAs(id?: string): VAAsResponse[] {
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
          `/api/vaas-sans-pythnet`
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
export default useLatestNonPythNetVAAs;
