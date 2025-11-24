import { Module, forwardRef } from '@nestjs/common';
import { EthersService } from './services/ethers.service';
import { AuthModule } from '../auth/auth.module';
import { DBModule } from '../db/db.module';

@Module({
  imports: [forwardRef(() => AuthModule), DBModule],
  controllers: [],
  providers: [EthersService],
  exports: [EthersService],
})
export class Web3Module {}
