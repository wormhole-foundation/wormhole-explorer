import { expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common';
import { EVMWatcher } from '../EVMWatcher';

jest.setTimeout(60000);

const initialOptimismBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.optimism);

test('getFinalizedBlockNumber', async () => {
  const watcher = new EVMWatcher('optimism');
  const blockNumber = await watcher.getFinalizedBlockNumber();
  // console.log('blockNumber', blockNumber);
  expect(blockNumber).toBeGreaterThan(105235062);
});

test('getMessagesForBlocks', async () => {
  const watcher = new EVMWatcher('optimism');
  const vaasByBlock = await watcher.getMessagesForBlocks(105235070, 105235080);
  expect(vaasByBlock).toMatchObject({
    '105235070/2023-06-06T16:28:37.000Z': [],
    '105235071/2023-06-06T16:28:39.000Z': [],
    '105235072/2023-06-06T16:28:41.000Z': [],
    '105235073/2023-06-06T16:28:43.000Z': [],
    '105235074/2023-06-06T16:28:45.000Z': [],
    '105235075/2023-06-06T16:28:47.000Z': [],
    '105235076/2023-06-06T16:28:49.000Z': [],
    '105235077/2023-06-06T16:28:51.000Z': [],
    '105235078/2023-06-06T16:28:53.000Z': [],
    '105235079/2023-06-06T16:28:55.000Z': [],
    '105235080/2023-06-06T16:28:57.000Z': [],
  });
});
