import { loadSsm } from '../util';
import { AwsConfigs, Config } from '../constant/type';

export const config: Config = {};

export const loadSsmConfig = async (awsConfig: AwsConfigs) => {
  if (awsConfig.param) {
    const ssmKeys: any = await loadSsm(awsConfig);

    if (!config.auth) config.auth = {};
    config.auth.hash = ssmKeys.AUTH.HASH;

    if (!config.aws) config.aws = {};
    if (!config.aws.s3) config.aws.s3 = {};
    config.aws.s3.bucket = ssmKeys.AWS.S3.BUCKET;
    config.aws.s3.accessKey = ssmKeys.AWS.S3.ACCESS_KEY_ID;
    config.aws.s3.secretKey = ssmKeys.AWS.S3.SECRET_ACCESS_KEY;

    if (!config.firebase) config.firebase = {};
    config.firebase.project_id = ssmKeys.FIREBASE.PROJECT_ID;
    config.firebase.client_email = ssmKeys.FIREBASE.CLIENT_EMAIL;
    config.firebase.private_key = ssmKeys.FIREBASE.PRIVATE_KEY;
    config.firebase.databaseURL = ssmKeys.FIREBASE.DATABASE_URL;

    if (!config.blockchain) config.blockchain = {};
    if (!config.blockchain.privateKey) config.blockchain.privateKey = {};
    config.blockchain.privateKey.owner = ssmKeys.BLOCKCHAIN.PRIVATE_KEY.OWNER;
    config.blockchain.privateKey.relayer =
      ssmKeys.BLOCKCHAIN.PRIVATE_KEY.RELAYER;
    config.blockchain.privateKey.relayer2 =
      ssmKeys.BLOCKCHAIN.PRIVATE_KEY.RELAYER2;
    config.blockchain.privateKey.relayer3 =
      ssmKeys.BLOCKCHAIN.PRIVATE_KEY.RELAYER3;
  }
};

export default () => ({
  ...config,
});
