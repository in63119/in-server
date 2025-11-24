import { Module, forwardRef } from '@nestjs/common';
import { AuthController } from './controllers/auth.controller';
import { AuthService } from './services/auth.service';
import { AccessJwtModule } from './jwt';
import { Web3Module } from '../web3/web3.module';

@Module({
  imports: [AccessJwtModule, forwardRef(() => Web3Module)],
  controllers: [AuthController],
  providers: [AuthService],
  exports: [AccessJwtModule, AuthService],
})
export class AuthModule {}
