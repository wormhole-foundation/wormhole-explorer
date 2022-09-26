import { HeartbeatNetwork, HeartbeatResponse } from "./useHeartbeats";

export type HeartbeatInfo = {
  guardian: string;
  name: string;
  network: HeartbeatNetwork;
};

export type ChainIdToHeartbeats = {
  [chainId: number]: HeartbeatInfo[];
};

function useChainHeartbeats(heartbeats: HeartbeatResponse[]) {
  const chainIdsToHeartbeats: ChainIdToHeartbeats = {};
  heartbeats.forEach((heartbeat) => {
    heartbeat.networks.forEach((network) => {
      if (!chainIdsToHeartbeats[network.id]) {
        chainIdsToHeartbeats[network.id] = [];
      }
      chainIdsToHeartbeats[network.id].push({
        guardian: heartbeat.guardianaddr,
        name: heartbeat.nodename || "",
        network,
      });
    });
  });
  return chainIdsToHeartbeats;
}
export default useChainHeartbeats;
