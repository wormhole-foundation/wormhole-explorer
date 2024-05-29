import { WormchainRepository } from "../../repositories";
import winston from "winston";
import { IbcTransaction, WormchainBlockLogs, CosmosRedeem } from "../../entities/wormchain";

const ATTRIBUTES_TYPES = ["wasm", "send_packet"];

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
      const wormchainLogs = await this.blockRepo.getBlockLogs(
        opts.chainId,
        blockNumber,
        ATTRIBUTES_TYPES
      );

      if (wormchainLogs && wormchainLogs.transactions && wormchainLogs.transactions.length > 0) {
        const ibcTransactions = this.findIbcTransactions(opts.addresses, wormchainLogs);

        if (ibcTransactions && ibcTransactions.length > 0) {
          const cosmosRedeems = await Promise.all(
            ibcTransactions.map((tx) => this.blockRepo.getRedeems(tx))
          );
          collectCosmosRedeems.push(...cosmosRedeems.flat());
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

  /*
   * This function parsing the wormchain logs.attributes to find the `cosmos transactions`
   * if we map packet_sequence, packet_timeout_timestamp, packet_src_channel, packet_dst_channel and targetChain
   * then we can consider it as a cosmos transaction and we can search for the `redeem` event for that transaction on cosmos chain
   */
  private findIbcTransactions(
    addresses: string[],
    wormchainLogs: WormchainBlockLogs
  ): IbcTransaction[] {
    const ibcTransactions: IbcTransaction[] = [];

    wormchainLogs.transactions?.forEach((tx) => {
      let gatewayContract: string | undefined;
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
              gatewayContract = value.toLowerCase();
            }
            break;
          case "transfer_payload":
            const valueDecoded = Buffer.from(attr.value, "base64").toString();
            const payload = Buffer.from(valueDecoded, "base64").toString();
            const payloadParsed = JSON.parse(payload) as GatewayTransfer;
            targetChain = payloadParsed.gateway_transfer.chain; // chain (osmosis, kujira, injective, evmos etc)
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
        gatewayContract &&
        targetChain &&
        srcChannel &&
        dstChannel &&
        timestamp &&
        sequence &&
        sender &&
        receiver
      ) {
        ibcTransactions.push({
          blockTimestamp: wormchainLogs.timestamp,
          gatewayContract,
          hash: tx.hash,
          targetChain,
          srcChannel,
          dstChannel,
          timestamp,
          tx: tx.tx,
          receiver,
          sequence,
          sender,
        });
      }
    });

    return ibcTransactions;
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
