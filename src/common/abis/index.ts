import type { InterfaceAbi } from 'ethers';

import AuthStorageDev from './kaia/test/development/AuthStorage.json';
import AuthStorageProd from './kaia/test/production/AuthStorage.json';
import PostStorageDev from './kaia/test/development/PostStorage.json';
import PostStorageProd from './kaia/test/production/PostStorage.json';
import PostForwarderDev from './kaia/test/development/PostForwarder.json';
import PostForwarderProd from './kaia/test/production/PostForwarder.json';
import VisitorStorageDev from './kaia/test/development/VisitorStorage.json';
import VisitorStorageProd from './kaia/test/production/VisitorStorage.json';
import YoutubeStorageDev from './kaia/test/development/YoutubeStorage.json';
import YoutubeStorageProd from './kaia/test/production/YoutubeStorage.json';
import SubscriberStorageDev from './kaia/test/development/SubscriberStorage.json';
import SubscriberStorageProd from './kaia/test/production/SubscriberStorage.json';
import RelayerManagerDev from './kaia/test/development/RelayerManager.json';
import RelayerManagerProd from './kaia/test/production/RelayerManager.json';

const SUPPORTED_ENVS = ['development', 'production'] as const;
type SupportedEnv = (typeof SUPPORTED_ENVS)[number];

const envInput = process.env.ENV;
const resolvedEnv =
  SUPPORTED_ENVS.find((value) => value === envInput) ?? 'development';

type ContractArtifact = {
  address: string;
  abi: InterfaceAbi;
};

type EnvArtifacts = {
  AuthStorage: ContractArtifact;
  PostStorage: ContractArtifact;
  PostForwarder: ContractArtifact;
  VisitorStorage: ContractArtifact;
  YoutubeStorage: ContractArtifact;
  SubscriberStorage: ContractArtifact;
  RelayerManager: ContractArtifact;
};

const abis: Record<SupportedEnv, EnvArtifacts> = {
  development: {
    AuthStorage: AuthStorageDev as ContractArtifact,
    PostStorage: PostStorageDev as ContractArtifact,
    PostForwarder: PostForwarderDev as ContractArtifact,
    VisitorStorage: VisitorStorageDev as ContractArtifact,
    YoutubeStorage: YoutubeStorageDev as ContractArtifact,
    SubscriberStorage: SubscriberStorageDev as ContractArtifact,
    RelayerManager: RelayerManagerDev as ContractArtifact,
  },
  production: {
    AuthStorage: AuthStorageProd as ContractArtifact,
    PostStorage: PostStorageProd as ContractArtifact,
    PostForwarder: PostForwarderProd as ContractArtifact,
    VisitorStorage: VisitorStorageProd as ContractArtifact,
    YoutubeStorage: YoutubeStorageProd as ContractArtifact,
    SubscriberStorage: SubscriberStorageProd as ContractArtifact,
    RelayerManager: RelayerManagerProd as ContractArtifact,
  },
};

export const getAbis = (): EnvArtifacts => {
  return abis[resolvedEnv];
};
