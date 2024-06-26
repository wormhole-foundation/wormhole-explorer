import { mapChain } from "../../../common/wormchain";
import { WormchainRedeem } from "../../entities/sei";
import { IbcTransaction, WormchainBlockLogs, CosmosRedeem } from "../../entities/wormchain";
import { WormchainRepository } from "../../repositories";
import { GetWormchainOpts } from "./PollWormchain";
import winston from "winston";

export class GetWormchainRedeems {
  private readonly blockRepo: WormchainRepository;
  protected readonly logger: winston.Logger;

  private previousFrom?: bigint;
  private lastFrom?: bigint;

  constructor(blockRepo: WormchainRepository) {
    this.logger = winston.child({ module: "GetWormchainRedeems" });
    this.blockRepo = blockRepo;
  }

  async execute(opts: GetWormchainOpts): Promise<CosmosRedeem[]> {
    const { chainId, addresses, blockBatchSize, previousFrom } = opts;
    const chain = mapChain(chainId);
    const collectCosmosRedeems: CosmosRedeem[] = [];

    this.logger.info(
      `[${chain}][exec] Processing range [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const wormchainRedeems = await this.blockRepo.getTxs(chainId, addresses[0], blockBatchSize);
    if (wormchainRedeems.length === 0) {
      return [];
    }

    const newLastFrom = BigInt(wormchainRedeems[wormchainRedeems.length - 1].height);
    if (previousFrom === newLastFrom) {
      return [];
    }

    const filteredWormchainRedeems =
      previousFrom && newLastFrom
        ? wormchainRedeems.filter(
            (cosmosRedeem) =>
              cosmosRedeem.height >= previousFrom && cosmosRedeem.height <= newLastFrom
          )
        : wormchainRedeems;

    await Promise.all(
      filteredWormchainRedeems.map(async (wormchainRedeem) => {
        const timestamp = await this.blockRepo.getBlockTimestamp(
          chainId,
          BigInt(wormchainRedeem.height)
        );
        wormchainRedeem.timestamp = timestamp;
      })
    );

    if (filteredWormchainRedeems && filteredWormchainRedeems.length > 0) {
      const ibcTransactions = this.findIbcTransactions(opts.addresses, filteredWormchainRedeems);

      if (ibcTransactions && ibcTransactions.length > 0) {
        const cosmosRedeems = await Promise.all(
          ibcTransactions.map((tx) => this.blockRepo.getRedeems(tx))
        );
        collectCosmosRedeems.push(...cosmosRedeems.flat());
      }
    }

    // Update previousFrom and lastFrom with opts lastFrom
    this.previousFrom = BigInt(wormchainRedeems[wormchainRedeems.length - 1].height);
    this.lastFrom = newLastFrom;

    this.logger.info(
      `[${chain}][exec] Got ${
        filteredWormchainRedeems?.length
      } transactions to process for ${this.populateLog(opts, this.previousFrom, this.lastFrom)}`
    );
    return filteredWormchainRedeems;
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
    filteredWormchainRedeems: WormchainRedeem[]
  ): IbcTransaction[] {
    const ibcTransactions: IbcTransaction[] = [];

    filteredWormchainRedeems?.forEach((tx) => {
      let gatewayContract: string | undefined;
      let targetChain: number | undefined;
      let srcChannel: string | undefined;
      let dstChannel: string | undefined;
      let timestamp: string | undefined;
      let receiver: string | undefined;
      let sequence: number | undefined;
      let sender: string | undefined;

      for (const event of tx.events) {
        for (const attr of event.attributes) {
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
          blockTimestamp: Number(timestamp),
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
