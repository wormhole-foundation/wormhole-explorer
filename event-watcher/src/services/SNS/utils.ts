import { env } from '../../config';
import AwsSNS from './AwsSNS';
import { AwsSNSConfig, SNSOptionTypes } from './types';

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
