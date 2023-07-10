import { CHAIN_ID_TO_NAME, ChainId } from "@certusone/wormhole-sdk";
import {
  getEnvironment,
  getRpcs,
  getSupportedChains,
  getWormholeRelayerAddressWrapped,
} from "./environment";
import { WormholeRelayer__factory } from "@certusone/wormhole-sdk/lib/cjs/ethers-contracts";
import { WebSocketProvider } from "./websocket";
import deliveryEventHandler from "./handlers/deliveryEventHandler";
import sendEventHandler from "./handlers/sendEventHandler";
import { EventHandler, getEventListener } from "./handlers/EventHandler";
import { Contract, ContractFactory, utils } from "ethers";

const ALL_EVENTS: EventHandler<any>[] = [
  deliveryEventHandler,
  sendEventHandler,
];

async function subscribeToEvents(chainId: ChainId, rpc: string) {
  console.log(`Subscribing to events for chain ${chainId}`);
  const ENVIRONMENT = await getEnvironment();

  for (const event of ALL_EVENTS) {
    if (event.shouldSupportChain(ENVIRONMENT, chainId)) {
      const contractAddress = event.getContractAddressEvm(ENVIRONMENT, chainId);
      const eventSignature = event.getEventSignatureEvm();
      if (!eventSignature) {
        continue;
      }
      const listener = getEventListener(event, chainId);
      const provider = new WebSocketProvider(rpc);
      provider.off(
        {
          address: contractAddress,
          topics: [utils.id(eventSignature)],
        },
        listener
      );
      provider.on(
        {
          address: contractAddress,
          topics: [utils.id(eventSignature)],
        },
        listener
      );
    }
  }

  console.log(`Subscribed to all events for chain ${chainId}`);
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
