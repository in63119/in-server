export enum CONTRACT_NAME {
  POSTSTORAGE = 'PostStorage',
  AUTHSTORAGE = 'AuthStorage',
  POSTFORWARDER = 'PostForwarder',
  RELAYERMANAGER = 'RelayerManager',
  VISITORSTORAGE = 'VisitorStorage',
  YOUTUBESTORAGE = 'YoutubeStorage',
  SUBSCRIBERSTORAGE = 'SubscriberStorage',
}

export enum RELAYER_STATUS {
  Ready = 'Ready',
  Processing = 'Processing',
  Shutdown = 'Shutdown',
}

export enum RELAYER_NUMBER {
  '0x74CA566E800b7FF5e5c5042FbA6ea31239B2Dea8' = 'relayer',
  '0x72caea34a00b7640282081827202fb40d89e1008' = 'relayer2',
  '0x077155fc3b07c0371ad8aa3a6f78701c71392366' = 'relayer3',
}
