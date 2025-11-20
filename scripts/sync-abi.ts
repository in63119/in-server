import path from 'path';
import fs from 'fs';
import {
  GetObjectCommand,
  S3Client,
  paginateListObjectsV2,
} from '@aws-sdk/client-s3';
import config, { loadSsmConfig } from '../src/common/config/default.config';
import dotenv from 'dotenv';

dotenv.config({ path: '.env' });

const abisPrefix = 'abis/';
const localBasePath = path.join(process.cwd(), '/src/common');

type S3Context = {
  client: S3Client;
  bucketName: string;
};

const env = process.env.ENV;
const awsRegion = process.env.AWS_REGION ?? '';
const awsSsmServer = process.env.AWS_SSM_SERVER ?? '';
const awsAccessKey = process.env.AWS_SSM_ACCESS_KEY ?? '';
const awsSecretKey = process.env.AWS_SSM_SECRET_KEY ?? '';

const resolveS3Context = (() => {
  let promise: Promise<S3Context> | null = null;

  return () => {
    if (!promise) {
      promise = (async () => {
        await loadSsmConfig({
          accessKey: awsAccessKey,
          secretAccessKey: awsSecretKey,
          region: awsRegion,
          param: `${awsSsmServer}/${env}`,
        });

        if (!config) {
          throw new Error('Configuration not loaded');
        }

        if (!config.aws) config.aws = {};
        if (!config.aws.s3) config.aws.s3 = {};

        const bucketName = config.aws?.s3?.bucket ?? '';
        const accessKeyId = config.aws?.s3?.accessKey ?? '';
        const secretAccessKey = config.aws?.s3?.secretKey ?? '';

        if (!bucketName) {
          throw new Error('AWS_S3_BUCKET is not configured');
        }
        if (!awsRegion) {
          throw new Error('AWS_REGION is not configured');
        }
        if (!accessKeyId || !secretAccessKey) {
          throw new Error('AWS credentials are not configured');
        }

        const client = new S3Client({
          region: awsRegion,
          credentials: {
            accessKeyId,
            secretAccessKey,
          },
        });

        return { client, bucketName };
      })();
    }

    return promise;
  };
})();

const syncAbis = async () => {
  try {
    const { client, bucketName } = await resolveS3Context();

    const objects = await listObjects(client, bucketName, abisPrefix);
    if (!objects || objects.length === 0) {
      console.log('No ABI files found in S3.');
      return;
    }

    for (const key of objects) {
      try {
        if (!key) continue;
        if (key.endsWith('/')) continue;

        const fullPath = path.join(localBasePath, key);
        const abiPath = path.dirname(fullPath);

        fs.mkdirSync(abiPath, { recursive: true });

        const abi = await getObject(client, bucketName, key);
        if (typeof abi !== 'string') continue;

        fs.writeFileSync(fullPath, abi);

        console.log(`Saved ABI file: ${fullPath}`);
      } catch (error) {
        console.error(error);
        continue;
      }
    }
  } catch (error) {
    console.error(error);
  }
};

const listObjects = async (
  client: S3Client,
  bucketName: string,
  prefix: string,
) => {
  const objects: string[] = [];

  try {
    const paginator = paginateListObjectsV2(
      { client },
      { Bucket: bucketName, Prefix: prefix },
    );

    for await (const page of paginator) {
      const keys = page.Contents?.map((o) => o.Key).filter(
        (k): k is string => !!k,
      );

      if (keys && keys.length > 0) {
        objects.push(...keys);
      }
    }

    return objects;
  } catch (error) {
    console.error(error);
  }
};

const getObject = async (client: S3Client, bucketName: string, key: string) => {
  try {
    const response = await client.send(
      new GetObjectCommand({
        Bucket: bucketName,
        Key: key,
      }),
    );
    if (!response.Body) {
      console.log('File not found');
      return;
    }

    const str = await response.Body.transformToString();
    return str;
  } catch (error) {
    console.error(error);
  }
};

syncAbis();
