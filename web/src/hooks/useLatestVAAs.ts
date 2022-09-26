import axios from "axios";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";

export type VAAsResponse = {
  createdAt: string;
  updatedAt: string;
  vaa: string;
  _id: string;
};

function useLatestVAAs(): VAAsResponse[] {
  const { currentNetwork } = useNetworkContext();
  const [vaas, setVAAs] = useState<VAAsResponse[]>([]);
  useEffect(() => {
    setVAAs([]);
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get<VAAsResponse[]>("/api/vaas");
        if (!cancelled) {
          setVAAs(response.data);
          await new Promise((resolve) => setTimeout(resolve, 1000));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork]);
  return vaas;
}
export default useLatestVAAs;
