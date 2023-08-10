import { expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common';
import { MoonbeamWatcher } from '../MoonbeamWatcher';

jest.setTimeout(60000);

const initialMoonbeamBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.moonbeam);

test('getFinalizedBlockNumber', async () => {
  const watcher = new MoonbeamWatcher();
  const blockNumber = await watcher.getFinalizedBlockNumber();
  expect(blockNumber).toBeGreaterThan(initialMoonbeamBlock);
});
