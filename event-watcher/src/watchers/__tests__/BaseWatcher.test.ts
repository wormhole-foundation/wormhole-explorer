import { expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common';
import { EVMWatcher } from '../EVMWatcher';

jest.setTimeout(60000);

const initialBaseBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.base);

test('getFinalizedBlockNumber', async () => {
  const watcher = new EVMWatcher('base');
  const blockNumber = await watcher.getFinalizedBlockNumber();
  // console.log('blockNumber', blockNumber);
  expect(blockNumber).toBeGreaterThan(initialBaseBlock);
});

test('getMessagesForBlocks', async () => {
  const watcher = new EVMWatcher('base');
  const vaasByBlock = await watcher.getMessagesForBlocks(1544175, 1544185);
  expect(vaasByBlock).toMatchObject({
    '1544175/2023-07-20T18:28:17.000Z': [],
    '1544176/2023-07-20T18:28:19.000Z': [],
    '1544177/2023-07-20T18:28:21.000Z': [],
    '1544178/2023-07-20T18:28:23.000Z': [],
    '1544179/2023-07-20T18:28:25.000Z': [],
    '1544180/2023-07-20T18:28:27.000Z': [],
    '1544181/2023-07-20T18:28:29.000Z': [],
    '1544182/2023-07-20T18:28:31.000Z': [],
    '1544183/2023-07-20T18:28:33.000Z': [],
    '1544184/2023-07-20T18:28:35.000Z': [],
    '1544185/2023-07-20T18:28:37.000Z': [],
  });
});

test('getMessagesForBlockWithWHMsg', async () => {
  const watcher = new EVMWatcher('base');
  const vaasByBlock = await watcher.getMessagesForBlocks(1557420, 1557429);
  expect(vaasByBlock).toMatchObject({
    '1557420/2023-07-21T01:49:47.000Z': [],
    '1557421/2023-07-21T01:49:49.000Z': [],
    '1557422/2023-07-21T01:49:51.000Z': [],
    '1557423/2023-07-21T01:49:53.000Z': [],
    '1557424/2023-07-21T01:49:55.000Z': [],
    '1557425/2023-07-21T01:49:57.000Z': [
      '0x9d217269dff740e74d21d32babbefe4bece7b88870b020f5505d3de3c6e59694:30/000000000000000000000000e2e2d9e31d7e1cc1178fe0d1c5950f6c809816a3/0',
    ],
    '1557426/2023-07-21T01:49:59.000Z': [],
    '1557427/2023-07-21T01:50:01.000Z': [],
    '1557428/2023-07-21T01:50:03.000Z': [],
    '1557429/2023-07-21T01:50:05.000Z': [],
  });
});
