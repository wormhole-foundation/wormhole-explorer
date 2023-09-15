import { env } from '../../config';
import AwsSNS from './AwsSNS';
import { AwsSNSConfig, SNSOptionTypes, SNSMessage } from './types';
import { WHTransaction } from '../../databases/types';
import crypto from 'node:crypto';

const AwsConfig: AwsSNSConfig = {
  region: env.AWS_SNS_REGION as string,
  subject: env.AWS_SNS_SUBJECT as string,
  topicArn: env.AWS_SNS_TOPIC_ARN as string,
  credentials: {
    accessKeyId: env.AWS_ACCESS_KEY_ID as string,
    secretAccessKey: env.AWS_SECRET_ACCESS_KEY as string,
  },
};

export const getSNS = (): SNSOptionTypes => {
  if (env.SNS_SOURCE === 'aws') return new AwsSNS(AwsConfig);
  return null;
};

export const makeSnsMessage = (
  whTx: WHTransaction,
  metadata: { source: string; type: string },
): SNSMessage => {
  const { id, eventLog } = whTx;
  const { emitterChain, emitterAddr, sequence, unsignedVaa, txHash, createdAt } = eventLog;
  const timestamp = createdAt ? new Date(createdAt).toISOString() : new Date().toISOString();
  const uuid = crypto.randomUUID();
  const trackId = `chain-event-${id}-${uuid}`;

  const snsMessage: SNSMessage = {
    trackId,
    source: metadata.source,
    type: metadata.type,
    payload: {
      id,
      emitterChain,
      emitterAddr,
      sequence,
      timestamp,
      vaa: unsignedVaa.toString('base64'),
      txHash: txHash,
    },
  };

  return snsMessage;
};
