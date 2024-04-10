import winston from "winston";
import { solana, LogFoundEvent } from "../../../domain/entities";
import base32 from "hi-base32";
import { IDL } from "./idls/tokenDispenser";
import { BorshCoder, EventParser, web3 } from "@coral-xyz/anchor";

const logger = winston.child({ module: "solanaAirdropInfluxMapper" });
const coder = new BorshCoder(IDL as any);
const eventParsers: Map<string, EventParser> = new Map();

export interface ProgramParams {
  instructions: string[];
  vaaAccountIndex: number;
}

export type Opts = {
  program: string;
};

export const solanaAirdropInfluxMapper = async (
  tx: solana.Transaction,
  { program }: Opts
): Promise<LogFoundEvent<{ tags: any; fields: any }>[]> => {
  if (!tx || !tx.blockTime) {
    throw new Error(
      `Block time is missing for tx ${tx?.transaction?.signatures} in slot ${tx?.slot}`
    );
  }
  if (!tx.meta || tx.meta?.err) {
    logger.silly(`Got a failed tx: ${tx.transaction.signatures[0]}`);

    return [asError(tx)];
  }

  if (!eventParsers.get(program)) {
    eventParsers.set(program, new EventParser(new web3.PublicKey(program), coder));
  }

  const parser = eventParsers.get(program)!;

  const eventGen = parser.parseLogs(tx.meta?.logMessages ?? []);
  const events = [];
  let event = eventGen.next();
  // Note: should only have 1 event/claim per txn at most
  while (!event.done) {
    events.push(event.value.data as Record<string, any>);
    event = eventGen.next();
  }

  logger.silly(`Extracted ${events.length} events from tx ${tx.transaction.signatures[0]}`);

  if (!events.length) {
    logger.warn(`No events found for tx ${tx.transaction.signatures[0]}. Ignoring.`);
    return [];
  }

  return [asSuccess(tx, events[0])];
};

const asError = (tx: solana.Transaction) => ({
  name: "failed_txn_event_v1",
  address: "",
  blockHeight: BigInt(tx.slot.toString()),
  blockTime: tx.blockTime!, // there is a check before validating this exists
  chainId: tx.chainId,
  txHash: tx.transaction.signatures[0],
  attributes: {
    tags: {
      network: "mainnet-beta", // TODO: add to cfg
    },
    fields: {
      signature: tx.transaction.signatures[0],
      errorDetails: JSON.stringify({
        blockTime: tx.blockTime,
        slot: tx.slot,
        err: tx.meta?.err,
      }),
    },
  },
});

const asSuccess = (tx: solana.Transaction, event?: Record<string, any>) => {
  const claimInfo = formatClaimInfo(event?.claimInfo);
  return {
    name: "txn_event_v1",
    address: "",
    blockHeight: BigInt(tx.slot.toString()),
    blockTime: tx.blockTime!, // there is a check before validating this exists
    chainId: tx.chainId,
    txHash: tx.transaction.signatures[0],
    attributes: {
      tags: {
        ecosystem: claimInfo?.ecosystem ?? "unknown",
        network: "mainnet-beta",
      },
      fields: {
        claimant: event?.claimant,
        address: claimInfo?.address,
        signature: tx.transaction.signatures[0],
        amount: Number(claimInfo?.amount),
        eventDetails: JSON.stringify({ ...event, claimInfo: claimInfo }),
      },
    },
  };
};

const formatClaimInfo = (claimInfo?: Record<string, any>) => {
  if (!claimInfo) return undefined;

  if (claimInfo.identity.discord) {
    return {
      ecosystem: "discord",
      address: claimInfo.identity.discord.username,
      amount: claimInfo.amount.toString(),
    };
  } else if (claimInfo.identity.solana) {
    return {
      ecosystem: "solana",
      address: new web3.PublicKey(claimInfo.identity.solana.pubkey).toBase58(),
      amount: claimInfo.amount.toString(),
    };
  } else if (claimInfo.identity.evm) {
    return {
      ecosystem: "evm",
      address: "0x" + Buffer.from(claimInfo.identity.evm.pubkey).toString("hex"),
      amount: claimInfo.amount.toString(),
    };
  } else if (claimInfo.identity.aptos) {
    return {
      ecosystem: "aptos",
      address: "0x" + Buffer.from(claimInfo.identity.aptos.address).toString("hex"),
      amount: claimInfo.amount.toString(),
    };
  } else if (claimInfo.identity.sui) {
    return {
      ecosystem: "sui",
      address: "0x" + Buffer.from(claimInfo.identity.sui.address).toString("hex"),
      amount: claimInfo.amount.toString(),
    };
  } else if (claimInfo.identity.cosmwasm) {
    return {
      ecosystem: "cosmwasm",
      address: claimInfo.identity.cosmwasm.address,
      amount: claimInfo.amount.toString(),
    };
  } else if (claimInfo.identity.injective) {
    return {
      ecosystem: "injective",
      address: claimInfo.identity.injective.address,
      amount: claimInfo.amount.toString(),
    };
  } else if (claimInfo.identity.algorand) {
    return {
      ecosystem: "algorand",
      address: base32.encode(claimInfo.identity.algorand.pubkey),
      amount: claimInfo.amount.toString(),
    };
  } else throw new Error(`unknown identity type. ${JSON.stringify(claimInfo.identity)}}`);
};
