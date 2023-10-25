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

async function queryEvents(chainId: ChainId, rpc: string) {
  console.log(`Querying events for chain ${chainId}`);
  const ENVIRONMENT = await getEnvironment();

  for (const event of ALL_EVENTS) {
    if (!event.shouldSupportChain(ENVIRONMENT, chainId)) {
      continue;
    }

    const contractAddress = event.getContractAddressEvm(ENVIRONMENT, chainId);
    const abi = event.getEventAbiEvm();
    const eventSignature = event.getEventSignatureEvm();
    const listener = getEventListener(event, chainId);

    if (!abi || !eventSignature) {
      continue;
    }

    const provider = new WebSocketProvider(rpc);
    const contract = new Contract(contractAddress, abi, provider);
    const filter = contract.filters[eventSignature]();
    const logs = await contract.queryFilter(filter, -2048, "latest");

    for (const log of logs) {
      await listener(log);
    }
  }

  console.log(`Queried events for chain ${chainId}`);
}

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
      try {
        provider.off(
          {
            address: contractAddress,
            topics: [utils.id(eventSignature)],
          },
          listener
        );
      } catch (e) {
        //ignore, we just want to make sure we don't have multiple listeners
      }
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

async function listenerLoop(sleepMs: number) {
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
    }
    console.log(`Initialized connections, sleeping for ${sleepMs}ms`);
    await sleep(sleepMs);
  }
}

export async function queryLoop(periodMs: number) {
  console.log("Starting query loop");
  const supportedChains = await getSupportedChains();
  const rpcs = await getRpcs();
  let run = true;
  while (run) {
    for (const chainId of supportedChains) {
      try {
        const rpc = rpcs.get(chainId);
        if (!rpc) {
          throw new Error("RPC not found");
        }
        await queryEvents(chainId, rpc);
      } catch (e) {
        console.error(`Error subscribing to events for chain ${chainId}`);
        console.error(e);
      }
    }
    await sleep(periodMs);
  }
}

async function sleep(timeout: number) {
  return new Promise((resolve) => setTimeout(resolve, timeout));
}

// start the process
listenerLoop(300000);
//queryLoop(300000);
