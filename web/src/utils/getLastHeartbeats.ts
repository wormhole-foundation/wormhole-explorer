import { publicrpc } from "@certusone/wormhole-sdk-proto-web";
import { Network } from "../contexts/NetworkContext";
const { GrpcWebImpl, PublicRPCServiceClientImpl } = publicrpc;

export async function getLastHeartbeats(network: Network) {
  const rpc = new GrpcWebImpl(network.endpoint, {});
  const api = new PublicRPCServiceClientImpl(rpc);
  return await api.GetLastHeartbeats({});
}
