export type Config = {
  aws?: {
    s3?: {
      bucket?: string;
      accessKey?: string;
      secretKey?: string;
    };
  };
};
