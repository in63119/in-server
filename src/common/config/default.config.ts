import { loadSsm } from '../util';
import { AwsConfigs, Config } from '../constant/type';

export const config: Config = {};

export const loadSsmConfig = async (awsConfig: AwsConfigs) => {
  if (awsConfig.param) {
    const ssmKeys: any = await loadSsm(awsConfig);

    if (!config.aws) config.aws = {};
    if (!config.aws.s3) config.aws.s3 = {};

    config.aws.s3.bucket = ssmKeys.AWS.S3.BUCKET;
    config.aws.s3.accessKey = ssmKeys.AWS.S3.ACCESS_KEY_ID;
    config.aws.s3.secretKey = ssmKeys.AWS.S3.SECRET_ACCESS_KEY;
  }
};

export default config;
