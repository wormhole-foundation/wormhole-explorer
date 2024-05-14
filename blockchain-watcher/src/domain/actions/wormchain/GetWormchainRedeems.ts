import { CosmosRedeem, WormchainBlockLogs, WormchainTransaction } from "../../entities/wormchain";
import { WormchainRepository } from "../../repositories";
import { MsgExecuteContract } from "cosmjs-types/cosmwasm/wasm/v1/tx";
import { decodeTxRaw } from "@cosmjs/proto-signing";
import { parseVaa } from "@certusone/wormhole-sdk";
import { base64 } from "ethers/lib/utils";
import winston from "winston";

const MSG_EXECUTE_CONTRACT_TYPE_URL = "/cosmwasm.wasm.v1.MsgExecuteContract";

export class GetWormchainRedeems {
  private readonly blockRepo: WormchainRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: WormchainRepository) {
    this.logger = winston.child({ module: "GetWormchainRedeems" });
    this.blockRepo = blockRepo;
  }

  async execute(
    range: Range,
    opts: { addresses: string[]; chainId: number }
  ): Promise<CosmosRedeem[]> {
    let fromBlock = range.fromBlock;
    let toBlock = range.toBlock;

    const collectCosmosRedeems: CosmosRedeem[] = [];

    if (fromBlock > toBlock) {
      this.logger.info(
        `[wormchain][exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
      );
      return [];
    }

    for (let blockNumber = fromBlock; blockNumber <= toBlock; blockNumber++) {
      const wormchainLogs = await this.blockRepo.getBlockLogs(opts.chainId, blockNumber, [
        "wasm",
        "send_packet",
      ]);

      if (wormchainLogs && wormchainLogs.transactions && wormchainLogs.transactions.length > 0) {
        const wormchainTransactions = await this.findWormchainTransactions(
          opts.addresses,
          wormchainLogs
        );

        // TODO: Improve this implementation
        if (wormchainTransactions && wormchainTransactions.length > 0) {
          for (const tx of wormchainTransactions) {
            const cosmosRedeems = await this.blockRepo.getRedeems(tx);
            for (const redeem of cosmosRedeems) {
              collectCosmosRedeems.push(redeem);
            }
          }
        }
      }
    }

    this.logger.info(
      `[wormchain][exec] Got ${
        collectCosmosRedeems?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return collectCosmosRedeems;
  }

  private populateLog(opts: { addresses: string[] }, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][blocks:${fromBlock} - ${toBlock}]`;
  }

  private async findWormchainTransactions(
    addresses: string[],
    wormchainLogs: WormchainBlockLogs
  ): Promise<any[]> {
    const wormchainTransactions: WormchainTransaction[] = [];

    wormchainLogs.transactions?.forEach(async (tx) => {
      let coreContract: string | undefined;
      let targetChain: number | undefined;
      let srcChannel: string | undefined;
      let dstChannel: string | undefined;
      let timestamp: string | undefined;
      let receiver: string | undefined;
      let sequence: number | undefined;
      let sender: string | undefined;

      for (const attr of tx.attributes) {
        const key = Buffer.from(attr.key, "base64").toString().toLowerCase();
        const value = Buffer.from(attr.value, "base64").toString().toLowerCase();

        switch (key) {
          case "_contract_address":
          case "contract_address":
            if (addresses.includes(value.toLowerCase())) {
              coreContract = value.toLowerCase();
            }
            break;
          case "transfer_payload":
            const valueDecoded = Buffer.from(attr.value, "base64").toString();
            const payload = Buffer.from(valueDecoded, "base64").toString();
            const payloadParsed = JSON.parse(payload) as GatewayTransfer;
            targetChain = payloadParsed.gateway_transfer.chain;
            break;
          case "packet_src_channel":
            srcChannel = value;
            break;
          case "packet_dst_channel":
            dstChannel = value;
            break;
          case "packet_timeout_timestamp":
            timestamp = value;
            break;
          case "packet_sequence":
            sequence = Number(value);
            break;
          case "packet_data":
            const packetData = JSON.parse(value) as PacketData;
            sender = packetData.receiver;
            receiver = packetData.sender;
            break;
        }
      }

      if (
        coreContract &&
        targetChain &&
        srcChannel &&
        dstChannel &&
        timestamp &&
        sequence &&
        sender &&
        receiver
      ) {
        const decodedTx = decodeTxRaw(tx.tx);
        const message = decodedTx.body.messages.find(
          (tx) => tx.typeUrl === MSG_EXECUTE_CONTRACT_TYPE_URL
        );

        if (message) {
          const parsedMessage = MsgExecuteContract.decode(message.value);

          const instruction = JSON.parse(Buffer.from(parsedMessage.msg).toString());
          const base64Vaa = instruction?.complete_transfer_and_convert?.vaa;

          if (base64Vaa) {
            const vaa = parseVaa(base64.decode(base64Vaa));

            wormchainTransactions.push({
              vaaEmitterAddress: vaa.emitterAddress.toString("hex").toUpperCase(),
              vaaEmitterChain: vaa.emitterChain,
              vaaSequence: vaa.sequence,
              blockTimestamp: wormchainLogs.timestamp,
              hash: tx.hash,
              coreContract,
              targetChain,
              srcChannel,
              dstChannel,
              timestamp,
              receiver,
              sequence,
              sender,
            });
          }
        }
      }
    });

    return wormchainTransactions;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};

type PacketData = {
  sender: string;
  receiver: string;
};

type GatewayTransfer = {
  gateway_transfer: {
    recipient: string;
    chain: number;
    nonce: number;
  };
};
