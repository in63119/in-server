import { loadSsm } from '../util';
import { AwsConfigs, Config } from '../constant/type';

const config: Config = {};

export const loadSsmConfig = async (awsConfig: AwsConfigs) => {
  if (awsConfig.param) {
    const ssmKeys: any = await loadSsm(awsConfig);

    if (!config.system) config.system = {};

    config.system.port = ssmKeys.system.port;
  }
};

export default () => ({
  ...config,
});
