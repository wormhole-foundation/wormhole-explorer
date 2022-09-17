import { Heartbeat_Network } from "@certusone/wormhole-sdk-proto-web/lib/cjs/gossip/v1/gossip";
import { GetLastHeartbeatsResponse_Entry } from "@certusone/wormhole-sdk-proto-web/lib/cjs/publicrpc/v1/publicrpc";

export type HeartbeatInfo = {
  guardian: string;
  name: string;
  network: Heartbeat_Network;
};

export type ChainIdToHeartbeats = {
  [chainId: number]: HeartbeatInfo[];
};

function useChainHeartbeats(heartbeats: GetLastHeartbeatsResponse_Entry[]) {
  const chainIdsToHeartbeats: ChainIdToHeartbeats = {};
  heartbeats.forEach((heartbeat) => {
    heartbeat.rawHeartbeat?.networks.forEach((network) => {
      if (!chainIdsToHeartbeats[network.id]) {
        chainIdsToHeartbeats[network.id] = [];
      }
      chainIdsToHeartbeats[network.id].push({
        guardian: heartbeat.p2pNodeAddr,
        name: heartbeat.rawHeartbeat?.nodeName || "",
        network,
      });
    });
  });
  return chainIdsToHeartbeats;
}
export default useChainHeartbeats;
