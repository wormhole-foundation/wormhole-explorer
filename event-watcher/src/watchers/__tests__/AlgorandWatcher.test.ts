import { expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common';
import { AlgorandWatcher } from '../AlgorandWatcher';

jest.setTimeout(180000);

const initialAlgorandBlock = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.algorand);

test('getFinalizedBlockNumber', async () => {
  const watcher = new AlgorandWatcher();
  const blockNumber = await watcher.getFinalizedBlockNumber();
  expect(blockNumber).toBeGreaterThan(initialAlgorandBlock);
});

test('getMessagesForBlocks', async () => {
  const watcher = new AlgorandWatcher();
  const messages = await watcher.getMessagesForBlocks(25692450, 25692450);
  expect(messages).toMatchObject({ '25692450/2022-12-21T02:00:40.000Z': [] });
});

test('getMessagesForBlocks initial block', async () => {
  const watcher = new AlgorandWatcher();
  const messages = await watcher.getMessagesForBlocks(initialAlgorandBlock, initialAlgorandBlock);
  expect(messages).toMatchObject({
    '22931277/2022-08-19T15:10:48.000Z': [
      '2RBQLCETCLFV4F3PQ7IHEWVWQV3MCP4UM5S5OFZM23XMC2O2DJ6A:8/67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45/1',
    ],
  });
});

test('getMessagesForBlocks indexer pagination support', async () => {
  const watcher = new AlgorandWatcher();
  const messages = await watcher.getMessagesForBlocks(initialAlgorandBlock, 27069946);
  expect(Object.keys(messages).length).toEqual(420);
});

test('getMessagesForBlocks seq < 192', async () => {
  const watcher = new AlgorandWatcher();
  const messages = await watcher.getMessagesForBlocks(25428873, 25428873);
  expect(messages).toMatchObject({
    '25428873/2022-12-09T18:10:08.000Z': [
      'M6QPEZ42P5O23II7SCWZTNZ7MHBSOH6KUNAPMH5YL3XHGNTEFUSQ:8/67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45/191',
    ],
  });
});

test('getMessagesForBlocks seq = 192', async () => {
  const watcher = new AlgorandWatcher();
  const messages = await watcher.getMessagesForBlocks(25433218, 25433218);
  expect(messages).toMatchObject({
    '25433218/2022-12-09T22:40:55.000Z': [
      '3PJPDBGTQK6JXAQEJNOYFE4NLLKFMCTKRY5FYNAXSEBDO25XUUJQ:8/67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45/192',
    ],
  });
});

test('getMessagesForBlocks seq > 383', async () => {
  const watcher = new AlgorandWatcher();
  const messages = await watcher.getMessagesForBlocks(26856742, 26856742);
  expect(messages).toMatchObject({
    '26856742/2023-02-09T09:05:04.000Z': [
      'LJNYXPG5VLJNNTBLSZSHLZQ7XQWTSUPKGA7APVI53J3MAKHQN72Q:8/67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45/384',
    ],
  });
});

test('getMessagesForBlocks on known empty block', async () => {
  const watcher = new AlgorandWatcher();
  const messages = await watcher.getMessagesForBlocks(23761195, 23761195);
  expect(messages).toMatchObject({ '23761195/2022-09-28T21:42:30.000Z': [] });
});
