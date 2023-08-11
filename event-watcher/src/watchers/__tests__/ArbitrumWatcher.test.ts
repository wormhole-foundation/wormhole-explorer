import { expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common';
import { ArbitrumWatcher } from '../ArbitrumWatcher';

jest.setTimeout(60000);

const initialArbitrumBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.arbitrum);
const initialEthBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.ethereum);

test('getFinalizedBlockNumber', async () => {
  const watcher = new ArbitrumWatcher();
  const blockNumber = await watcher.getFinalizedBlockNumber();
  expect(blockNumber).toEqual(0);
  let retval: number[] = watcher.getFirstMapEntry();
  expect(retval[0]).toBeGreaterThan(initialEthBlock);
  expect(retval[1]).toBeGreaterThan(initialArbitrumBlock);
});

// The following test will be enabled when there is a block to see.
test.skip('getMessagesForBlocks', async () => {
  const watcher = new ArbitrumWatcher();
  const vaasByBlock = await watcher.getMessagesForBlocks(53473701, 53473701);
  const entries = Object.entries(vaasByBlock);
  expect(entries.length).toEqual(1);
  expect(entries.filter(([block, vaas]) => vaas.length === 0).length).toEqual(1);
  expect(entries.filter(([block, vaas]) => vaas.length === 1).length).toEqual(0);
  expect(entries.filter(([block, vaas]) => vaas.length === 2).length).toEqual(0);
  expect(vaasByBlock['10974196/2023-01-06T04:23:21.045Z']).toBeDefined();
  expect(vaasByBlock['10974196/2023-01-06T04:23:21.045Z'].length).toEqual(1);
  expect(vaasByBlock['10974196/2023-01-06T04:23:21.045Z'][0]).toEqual(
    '8A31CDE56ED3ACB7239D705949BD6C164747210A6C4C69D98756E0CF6D22C9EB:3/0000000000000000000000007cf7b764e38a0a5e967972c1df77d432510564e2/256813'
  );
});
