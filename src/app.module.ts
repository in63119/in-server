import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import config from './common/config/default.config';

import { AppController } from './app.controller';
import { AppService } from './app.service';

import { EmailModule, DBModule, AuthModule, Web3Module } from './modules';

@Module({
  imports: [
    ConfigModule.forRoot({ isGlobal: true, load: [config] }),
    EmailModule,
    DBModule,
    AuthModule,
    Web3Module,
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
