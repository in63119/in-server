import { SSMClient, GetParameterCommand } from '@aws-sdk/client-ssm';
import type { AwsConfigs } from '../constant/type';

export const loadSsm = async ({
  accessKey,
  secretAccessKey,
  region,
  param,
}: AwsConfigs) => {
  if (!accessKey || !secretAccessKey) {
    throw new Error('AWS accessKey or secretAccessKey not found');
  }

  const client = new SSMClient({
    region,
    credentials: {
      accessKeyId: accessKey,
      secretAccessKey: secretAccessKey,
    },
  });
  const command = new GetParameterCommand({
    Name: param,
    WithDecryption: false,
  });

  try {
    const { Parameter } = await client.send(command);
    if (!Parameter || typeof Parameter.Value !== 'string') {
      throw new Error(`SSM Parameter  not found`);
    }

    return JSON.parse(Parameter.Value);
  } catch (error) {
    console.error('ssm error', error);
  }
};
