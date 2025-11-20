import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { allowedOrigins } from './common/constant/origin';
import { loadSsmConfig } from './common/config/default.config';

const env = process.env.ENV || '';
if (env.length === 0) {
  throw new Error('ENV is not defined');
}

const rawPort = process.env.PORT;
if (!rawPort) {
  throw new Error('PORT is not defined');
}
const PORT = parseInt(rawPort, 10);
if (Number.isNaN(PORT)) {
  throw new Error('PORT is not a valid number');
}

const awsRegion = process.env.AWS_REGION ?? '';
const awsSsmServer = process.env.AWS_SSM_SERVER ?? '';
const awsAccessKey = process.env.AWS_SSM_ACCESS_KEY ?? '';
const awsSecretKey = process.env.AWS_SSM_SECRET_KEY ?? '';

async function bootstrap() {
  await loadSsmConfig({
    accessKey: awsAccessKey,
    secretAccessKey: awsSecretKey,
    region: awsRegion,
    param: `${awsSsmServer}/${env}`,
  });

  const app = await NestFactory.create(AppModule);

  app.enableCors({
    origin: [allowedOrigins.local, allowedOrigins.prod],
    credentials: true,
    methods: ['GET', 'POST', 'OPTIONS', 'DELETE', 'PUT', 'PATCH'],
  });

  await app.listen(PORT);
}

bootstrap();
