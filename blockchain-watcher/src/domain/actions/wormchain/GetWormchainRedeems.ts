import { CosmosRedeem, WormchainBlockLogs } from "../../entities/wormchain";
import { WormchainRepository } from "../../repositories";
import winston from "winston";

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
  ): Promise<WormchainBlockLogs[]> {
    let fromBlock = range.fromBlock;
    let toBlock = range.toBlock;

    const collectWormchainLogs: any[] = [];

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
        const cosmosRedeems = await this.filterRedeemsTransactions(opts.addresses, wormchainLogs);

        if (cosmosRedeems && cosmosRedeems.length > 0) {
          collectWormchainLogs.push(cosmosRedeems.forEach((redeem) => redeem)); // TODO: Improve this implementation
        }
      }
    }

    this.logger.info(
      `[wormchain][exec] Got ${
        collectWormchainLogs?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return collectWormchainLogs;
  }

  private populateLog(opts: { addresses: string[] }, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][blocks:${fromBlock} - ${toBlock}]`;
  }

  private async filterRedeemsTransactions(
    addresses: string[],
    wormchainLogs: WormchainBlockLogs
  ): Promise<any[]> {
    const cosmosRedeems: CosmosRedeem[] = [];

    wormchainLogs.transactions?.forEach(async (tx) => {
      let coreContract: string | undefined;
      let srcChannel: string | undefined;
      let dstChannel: string | undefined;
      let timestamp: string | undefined;
      let receiver: string | undefined;
      let sequence: number | undefined;
      let chainId: number | undefined;
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
            chainId = payloadParsed.gateway_transfer.chain;
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
        srcChannel &&
        chainId &&
        dstChannel &&
        timestamp &&
        sequence &&
        sender &&
        receiver
      ) {
        cosmosRedeems.push({
          height: tx.height,
          hash: tx.hash,
          coreContract,
          srcChannel,
          dstChannel,
          timestamp,
          receiver,
          sequence,
          chainId,
          sender,
        });
      }
    });

    return cosmosRedeems;
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

// Map fist the chainId and your respective rpc
//const u = `https://rpc-osmosis-ia.cosmosia.notional.ventures/tx_search?query="tx.height=${resultTransaction.result.height}"`;

//const test: any = await this.pool
//.get()
//.get(u);
