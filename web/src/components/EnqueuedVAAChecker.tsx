import { ChainId, getSignedVAA } from "@certusone/wormhole-sdk";
import { GovernorGetEnqueuedVAAsResponse_Entry } from "@certusone/wormhole-sdk-proto-web/lib/cjs/publicrpc/v1/publicrpc";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";

const VAA_CHECK_TIMEOUT = 60000;

function EnqueuedVAAChecker({
  vaa: { emitterAddress, emitterChain, sequence },
}: {
  vaa: GovernorGetEnqueuedVAAsResponse_Entry;
}) {
  const {
    currentNetwork: { endpoint },
  } = useNetworkContext();
  const [vaaHasQuorum, setVaaHasQuorum] = useState<boolean | null>(null);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        setVaaHasQuorum(null);
        let result = false;
        try {
          const response = await getSignedVAA(
            endpoint,
            emitterChain as ChainId,
            emitterAddress,
            sequence
          );
          if (!!response.vaaBytes) result = true;
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
  }, [endpoint, emitterChain, emitterAddress, sequence]);
  return (
    <span role="img">
      {vaaHasQuorum === null ? "⏳" : vaaHasQuorum ? "✅" : "❌"}
    </span>
  );
}

export default EnqueuedVAAChecker;
