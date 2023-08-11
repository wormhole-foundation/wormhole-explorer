import { expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common';
import { PolygonWatcher } from '../PolygonWatcher';

jest.setTimeout(60000);

const initialPolygonBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.polygon);

test('getFinalizedBlockNumber', async () => {
  const watcher = new PolygonWatcher();
  const blockNumber = await watcher.getFinalizedBlockNumber();
  expect(blockNumber).toBeGreaterThan(initialPolygonBlock);
});
