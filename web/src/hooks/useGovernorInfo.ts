import {
  GovernorGetAvailableNotionalByChainResponse_Entry,
  GovernorGetEnqueuedVAAsResponse_Entry,
  GovernorGetTokenListResponse_Entry,
} from "@certusone/wormhole-sdk-proto-web/lib/cjs/publicrpc/v1/publicrpc";
import { useEffect, useState } from "react";
import { useNetworkContext } from "../contexts/NetworkContext";
import { getGovernorAvailableNotionalByChain } from "../utils/getGovernorAvailableNotionalByChain";
import { getGovernorEnqueuedVAAs } from "../utils/getGovernorEnqueuedVAAs";
import { getGovernorTokenList } from "../utils/getGovernorTokenList";

type GovernorInfo = {
  notionals: GovernorGetAvailableNotionalByChainResponse_Entry[];
  tokens: GovernorGetTokenListResponse_Entry[];
  enqueued: GovernorGetEnqueuedVAAsResponse_Entry[];
};

const createEmptyInfo = (): GovernorInfo => ({
  notionals: [],
  tokens: [],
  enqueued: [],
});

const TIMEOUT = 10 * 1000;

function useGovernorInfo(): GovernorInfo {
  const { currentNetwork } = useNetworkContext();
  const [governorInfo, setGovernorInfo] = useState<GovernorInfo>(
    createEmptyInfo()
  );
  useEffect(() => {
    setGovernorInfo(createEmptyInfo());
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await getGovernorAvailableNotionalByChain(
          currentNetwork
        );
        if (!cancelled) {
          setGovernorInfo((info) => ({ ...info, notionals: response.entries }));
          await new Promise((resolve) => setTimeout(resolve, TIMEOUT));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      // TODO: only update GovernorInfo with changes to token list, but that will cause displaySymbols to break
      while (!cancelled) {
        const response = await getGovernorTokenList(currentNetwork);
        if (!cancelled) {
          setGovernorInfo((info) => ({ ...info, tokens: response.entries }));
          await new Promise((resolve) => setTimeout(resolve, TIMEOUT));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork]);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await getGovernorEnqueuedVAAs(currentNetwork);
        if (!cancelled) {
          setGovernorInfo((info) => ({ ...info, enqueued: response.entries }));
          await new Promise((resolve) => setTimeout(resolve, TIMEOUT));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentNetwork]);
  return governorInfo;
}
export default useGovernorInfo;
