import { grabTerraCustodyData } from "./getTerraCustody";
import { grabSolanaCustodyData } from "./getSolanaCustody";
import { grabEvmCustodyData } from "./getEvmCustody";
import { MongoClient } from "mongodb";
import { grabNearCustodyData } from "./getNearCustody";
import { grabAlgoCustodyData } from "./getAlgorandCustody";
import { grabAptosCustodyData } from "./getAptosCustody";
import { sleepFor } from "./utils";
interface Token {
  tokenAddress: string;
  name: string;
  decimals: number;
  symbol: string;
  balance: BigInt;
  qty: number;
  tokenPrice: number;
  tokenBalanceUSD: number;
}

interface CustodyInfo {
  _id: string;
  updatedAt: string;
  chainName: string;
  chainId: number;
  emitterAddress: string;
  custodyUSD: number;
  tokens: Token[];
}

async function updateTable(chainInfo, client: MongoClient) {
  if (chainInfo === undefined) {
    return;
  }
  const custodyList = chainInfo.balances;
  if (custodyList.length === 0) {
    return;
  }
  try {
    const totalCustodyUSD = custodyList
      .map((x) => x.tokenBalanceUSD)
      .reduce((partialSum, a) => partialSum + a, 0);
    console.log("totalCustodyUSD=", totalCustodyUSD);

    const database = client.db("onchain_data");
    // Specifying a Schema is optional, but it enables type hints on
    // finds and inserts
    const chainId = chainInfo.chain_id;
    const emitterAddress = chainInfo.emitter_address;
    const custody = database.collection<CustodyInfo>("custody");
    const result = await custody.updateOne(
      { _id: `${chainId}/${emitterAddress}` },
      {
        $set: {
          updatedAt: new Date().toISOString(),
          chainName: chainInfo.name,
          chainId: chainId,
          emitterAddress: emitterAddress,
          custodyUSD: totalCustodyUSD,
          tokens: custodyList,
          _id: `${chainId}/${emitterAddress}`,
        },
      },
      { upsert: true }
    );
    console.log(`A document was inserted with the _id: ${result.upsertedId}`);
  } catch (e) {
    console.log(encodeURIComponent);
  }
  return;
}

const useAllowListstr = process.env.allowlist || "false";

export async function getCustodyData() {
  const uri = process.env.MONGODB_URI;
  if (uri === "" || uri === undefined) {
    console.log("No mongodb uri supplied");
    return -1;
  }
  const client = new MongoClient(uri);

  const useAllowList = true ? useAllowListstr === "true" : false;
  const timeout = 5000;
  const promises = [
    grabSolanaCustodyData("1", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("2", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabTerraCustodyData("3", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("4", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("5", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("6", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("7", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabAlgoCustodyData("8", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("9", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("10", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("11", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("12", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("13", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("14", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabNearCustodyData("15", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabEvmCustodyData("16", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabTerraCustodyData("18", useAllowList),
    await new Promise((res) => setTimeout(res, timeout)),
    grabAptosCustodyData("22", useAllowList),
    // grabTerraCustodyData("28", useAllowList),
  ];

  const output = await Promise.all(promises);
  // iterate through chains & insert into mongodb
  try {
    for (let i = 0; i < output.length; i++) {
      const data = output[i];
      await updateTable(data, client);
    }
  } catch (e) {
    console.log(e);
  } finally {
    await client.close();
  }
}

export default getCustodyData;
