export type Config = {
  auth?: {
    hash?: string;
    jwt?: {
      accessSecret?: string;
    };
  };
  aws?: {
    s3?: {
      bucket?: string;
      accessKey?: string;
      secretKey?: string;
    };
  };
  blockchain?: {
    privateKey?: {
      owner?: string;
      relayer?: string;
      relayer2?: string;
      relayer3?: string;
    };
  };
  firebase?: {
    project_id?: string;
    client_email?: string;
    private_key?: string;
    databaseURL?: string;
  };
};
