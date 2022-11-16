import axios from "axios";
import { useEffect, useState } from "react";
import { EnqueuedVaaDetailsResponse } from "../hooks/useEnqueuedVaaDetails";

const VAA_CHECK_TIMEOUT = 60000;

function EnqueuedVaaExists(row: EnqueuedVaaDetailsResponse) {
  const [vaaHasQuorum, setVaaHasQuorum] = useState<boolean | null>(null);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        setVaaHasQuorum(null);
        let result = false;
        try {
          const response = await axios.get(
            `/api/vaas/${row.chainId}/${row.emitterAddress.slice(2)}/${
              row.sequence
            }`
          );
          if (response.data) result = true;
        } catch (e) {}
        if (!cancelled) {
          setVaaHasQuorum(result);
          if (result) {
            cancelled = true;
            return;
          }
          await new Promise((resolve) =>
            setTimeout(resolve, VAA_CHECK_TIMEOUT)
          );
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [row]);
  return (
    <span role="img">
      {vaaHasQuorum === null ? "⏳" : vaaHasQuorum ? "✅" : "❌"}
    </span>
  );
}

export default EnqueuedVaaExists;
