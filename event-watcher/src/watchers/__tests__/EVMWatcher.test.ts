import { CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { expect, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common';
import { Block, EVMWatcher, LOG_MESSAGE_PUBLISHED_TOPIC } from '../EVMWatcher';

const initialAvalancheBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.avalanche);
const initialCeloBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.celo);
const initialOasisBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.oasis);
const initialKaruraBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.karura);

test('getBlock by tag', async () => {
  const watcher = new EVMWatcher('avalanche');
  const block = await watcher.getBlock('latest');
  expect(block.number).toBeGreaterThan(initialAvalancheBlock);
  expect(block.timestamp).toBeGreaterThan(1671672811);
  expect(new Date(block.timestamp * 1000).toISOString() > '2022-12-21').toBeTruthy();
});

test('getBlock by number', async () => {
  const watcher = new EVMWatcher('avalanche');
  const block = await watcher.getBlock(initialAvalancheBlock);
  expect(block.number).toEqual(initialAvalancheBlock);
  expect(block.hash).toEqual('0x33b358fe68a2a11b6a5a5969f29f9223001e36a5d719734ba542b238d397f14e');
  expect(block.timestamp).toEqual(1639504758);
  expect(new Date(block.timestamp * 1000).toISOString()).toEqual('2021-12-14T17:59:18.000Z');
});

test('getBlocks', async () => {
  const watcher = new EVMWatcher('avalanche');
  const blocks = await watcher.getBlocks(
    initialAvalancheBlock,
    initialAvalancheBlock + watcher.maximumBatchSize - 1
  );
  expect(blocks.length).toEqual(watcher.maximumBatchSize);
  expect(blocks[0].number).toEqual(initialAvalancheBlock);
  expect(blocks[0].hash).toEqual(
    '0x33b358fe68a2a11b6a5a5969f29f9223001e36a5d719734ba542b238d397f14e'
  );
  expect(blocks[0].timestamp).toEqual(1639504758);
  expect(new Date(blocks[0].timestamp * 1000).toISOString()).toEqual('2021-12-14T17:59:18.000Z');
  expect(blocks[99].number).toEqual(initialAvalancheBlock + 99);
  expect(blocks[99].hash).toEqual(
    '0x598080458a28e1241528d0d8c745425147179b86e353d5b0e5cc29e4154d13f6'
  );
  expect(blocks[99].timestamp).toEqual(1639504940);
});

test('getLogs', async () => {
  const watcher = new EVMWatcher('avalanche');
  const logs = await watcher.getLogs(9743300, 9743399, CONTRACTS.MAINNET.avalanche.core, [
    LOG_MESSAGE_PUBLISHED_TOPIC,
  ]);
  expect(logs.length).toEqual(2);
  expect(logs[0].topics[0]).toEqual(LOG_MESSAGE_PUBLISHED_TOPIC);
  expect(logs[0].blockNumber).toEqual(9743306);
  expect(logs[0].transactionHash).toEqual(
    '0x0ca26f28b454591e600ff03fcff60e35bf74f12ebe0c3ba2165a6b6d5a5e4da8'
  );
});

test('getFinalizedBlockNumber', async () => {
  const watcher = new EVMWatcher('avalanche');
  const blockNumber = await watcher.getFinalizedBlockNumber();
  expect(blockNumber).toBeGreaterThan(initialAvalancheBlock);
});

test('getMessagesForBlocks', async () => {
  const watcher = new EVMWatcher('avalanche');
  const vaasByBlock = await watcher.getMessagesForBlocks(9743300, 9743399);
  const entries = Object.entries(vaasByBlock);
  expect(entries.length).toEqual(100);
  expect(entries.filter(([block, vaas]) => vaas.length === 0).length).toEqual(98);
  expect(entries.filter(([block, vaas]) => vaas.length === 1).length).toEqual(2);
  expect(entries.filter(([block, vaas]) => vaas.length === 2).length).toEqual(0);
  expect(vaasByBlock['9743306/2022-01-18T17:59:33.000Z']).toBeDefined();
  expect(vaasByBlock['9743306/2022-01-18T17:59:33.000Z'].length).toEqual(1);
  expect(vaasByBlock['9743306/2022-01-18T17:59:33.000Z'][0]).toEqual(
    '0x0ca26f28b454591e600ff03fcff60e35bf74f12ebe0c3ba2165a6b6d5a5e4da8:6/0000000000000000000000000e082f06ff657d94310cb8ce8b0d9a04541d8052/3683'
  );
});

test('getBlock by tag (Oasis compatibility)', async () => {
  const watcher = new EVMWatcher('oasis');
  const block = await watcher.getBlock('latest');
  expect(block.number).toBeGreaterThan(initialOasisBlock);
  expect(block.timestamp).toBeGreaterThan(3895665);
  expect(new Date(block.timestamp * 1000).toISOString() > '2022-12-21').toBeTruthy();
});

test('getBlock by tag (Celo compatibility)', async () => {
  const watcher = new EVMWatcher('celo');
  const block = await watcher.getBlock('latest');
  expect(block.number).toBeGreaterThan(initialCeloBlock);
  expect(block.timestamp).toBeGreaterThan(1671672811);
  expect(new Date(block.timestamp * 1000).toISOString() > '2022-12-21').toBeTruthy();
});

test('getBlock by number (Celo compatibility)', async () => {
  const watcher = new EVMWatcher('celo');
  const block = await watcher.getBlock(initialCeloBlock);
  expect(block.number).toEqual(initialCeloBlock);
  expect(block.timestamp).toEqual(1652314820);
  expect(new Date(block.timestamp * 1000).toISOString()).toEqual('2022-05-12T00:20:20.000Z');
});

test('getMessagesForBlocks (Celo compatibility)', async () => {
  const watcher = new EVMWatcher('celo');
  const vaasByBlock = await watcher.getMessagesForBlocks(13322450, 13322549);
  const entries = Object.entries(vaasByBlock);
  expect(entries.length).toEqual(100);
  expect(entries.filter(([block, vaas]) => vaas.length === 0).length).toEqual(98);
  expect(entries.filter(([block, vaas]) => vaas.length === 1).length).toEqual(2);
  expect(entries.filter(([block, vaas]) => vaas.length === 2).length).toEqual(0);
  expect(vaasByBlock['13322492/2022-06-02T17:40:22.000Z']).toBeDefined();
  expect(vaasByBlock['13322492/2022-06-02T17:40:22.000Z'].length).toEqual(1);
  expect(vaasByBlock['13322492/2022-06-02T17:40:22.000Z'][0]).toEqual(
    '0xd73c03b0d59ecae473d50b61e8756bc19b54314869e9b11d0fda6f89dbcf3918:14/000000000000000000000000796dff6d74f3e27060b71255fe517bfb23c93eed/5'
  );
});

test('getBlock by number (Karura compatibility)', async () => {
  const watcher = new EVMWatcher('karura');
  const latestBlock = await watcher.getFinalizedBlockNumber();
  const moreRecentBlockNumber = 4646601;
  //   block {
  //   hash: '0xe370a794f27fc49d1e468c78e4f92f9aeefc949a62f919cea8d2bd81904840b5',
  //   number: 4646601,
  //   timestamp: 1687963290
  // }
  expect(latestBlock).toBeGreaterThan(moreRecentBlockNumber);
  const block = await watcher.getBlock(moreRecentBlockNumber);
  expect(block.number).toEqual(moreRecentBlockNumber);
  expect(block.timestamp).toEqual(1687963290);
  expect(new Date(block.timestamp * 1000).toISOString()).toEqual('2023-06-28T14:41:30.000Z');
});

test('getMessagesForBlocks (Karura compatibility)', async () => {
  const watcher = new EVMWatcher('karura');
  const vaasByBlock = await watcher.getMessagesForBlocks(4582511, 4582513);
  const entries = Object.entries(vaasByBlock);
  // console.log('entries', entries);
  expect(entries.length).toEqual(3);
  expect(entries[0][0]).toEqual('4582511/2023-06-19T15:54:48.000Z');
  // 4582512 was an error block. In that case, make sure it has the same timestamp as the previous block
  // expect(entries[1][0]).toEqual('4582512/2023-06-19T15:54:48.000Z');
  // As of July 15, 2023, the above block appears to have been fixed
  expect(entries[1][0]).toEqual('4582512/2023-06-19T15:55:00.000Z');
});

test('getMessagesForBlocks (Karura compatibility 2)', async () => {
  const watcher = new EVMWatcher('karura');
  await watcher.getFinalizedBlockNumber(); // This has the side effect of initializing the latestFinalizedBlockNumber
  const vaasByBlock = await watcher.getMessagesForBlocks(4595356, 4595358);
  const entries = Object.entries(vaasByBlock);
  // console.log('entries', entries);
  expect(entries.length).toEqual(3);
});

test('getBlock (Karura compatibility)', async () => {
  const watcher = new EVMWatcher('karura');
  await watcher.getFinalizedBlockNumber(); // This has the side effect of initializing the latestFinalizedBlockNumber
  let block: Block = await watcher.getBlock(4582512); // 6969 block
  // console.log('block', block);
  block = await watcher.getBlock(4595357); // Null block
  // console.log('block', block);
  // block = await watcher.getBlock(4595358); // good block
  // console.log('block', block);
  // block = await watcher.getBlock(4619551); // good luck
  // console.log('block', block);
});
