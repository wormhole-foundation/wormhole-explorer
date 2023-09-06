import { env } from '../../config';
import AwsSNS from './AwsSNS';
import { AwsSNSConfig, SNSOptionTypes, SNSMessage } from './types';
import { VaaLog } from '../../databases/types';

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
  vaaLog: VaaLog,
  metadata: { source: string; type: string },
): SNSMessage => {
  const { trackId, id, chainId, emitter, sequence, txHash, payload, createdAt } = vaaLog;
  const timestamp = createdAt ? new Date(createdAt).toISOString() : new Date().toISOString();

  const snsMessage: SNSMessage = {
    trackId: trackId,
    source: metadata.source,
    type: metadata.type,
    payload: {
      id,
      emitterChain: chainId,
      emitterAddr: emitter,
      sequence,
      timestamp,
      vaa: payload,
      txHash: txHash,
    },
  };

  return snsMessage;
};
