import { CHAIN_ID_TO_NAME, ChainId } from "@certusone/wormhole-sdk";
import {
  getEnvironment,
  getRpcs,
  getSupportedChains,
  getWormholeRelayerAddressWrapped,
} from "./environment";
import { WormholeRelayer__factory } from "@certusone/wormhole-sdk/lib/cjs/ethers-contracts";
import { WebSocketProvider } from "./websocket";
import { handleSendEvent } from "./handlers/sendEventHandler";
import { handleDeliveryEvent } from "./handlers/deliveryEventHandler";

async function subscribeToEvents(chainId: ChainId, rpc: string) {
  console.log(`Subscribing to events for chain ${chainId}`);
  const ENVIRONMENT = await getEnvironment();

  const wormholeRelayerAddress = getWormholeRelayerAddressWrapped(
    CHAIN_ID_TO_NAME[chainId],
    ENVIRONMENT
  );
  if (!wormholeRelayerAddress) {
    throw new Error(
      `Wormhole Relayer contract address not found for chain ${chainId}`
    );
  }

  const wormholeRelayer = WormholeRelayer__factory.connect(
    wormholeRelayerAddress,
    new WebSocketProvider(rpc)
  );

  // unsubscribe to reset websocket connection
  wormholeRelayer.off("SendEvent(uint64,uint256,uint256)", (...args) => {
    // @ts-ignore
    return handleSendEvent(chainId, ...args);
  });
  wormholeRelayer.off(
    "Delivery(address,uint16,uint64,bytes32,uint8,uint256,uint8,bytes,bytes)",
    (...args) => {
      // @ts-ignore
      return handleDeliveryEvent(chainId, ...args);
    }
  );

  // resubscribe to events
  wormholeRelayer.on("SendEvent(uint64,uint256,uint256)", (...args) => {
    // @ts-ignore
    return handleSendEvent(chainId, ...args);
  });
  wormholeRelayer.on(
    "Delivery(address,uint16,uint64,bytes32,uint8,uint256,uint8,bytes,bytes)",
    (...args) => {
      // @ts-ignore
      return handleDeliveryEvent(chainId, ...args);
    }
  );

  console.log(
    `Subscribed to: ${chainId}, wormholeRelayer contract: ${wormholeRelayerAddress}`
  );
}

async function main(sleepMs: number) {
  console.log("Starting event watcher");
  const SUPPORTED_CHAINS = await getSupportedChains();

  let run = true;
  while (run) {
    // resubscribe to contract events every 5 minutes
    for (const chainId of SUPPORTED_CHAINS) {
      try {
        const rpc = (await getRpcs()).get(chainId);
        if (!rpc) {
          console.log(`RPC not found for chain ${chainId}`);
          //hard exit
          process.exit(1);
        }
        await subscribeToEvents(chainId, rpc);
      } catch (e: any) {
        console.log(e);
        run = false;
      }
      console.log(`Initialized connections, sleeping for ${sleepMs}ms`);
    }
    await sleep(sleepMs);
  }
}

async function sleep(timeout: number) {
  return new Promise((resolve) => setTimeout(resolve, timeout));
}

// start the process
main(300000);
