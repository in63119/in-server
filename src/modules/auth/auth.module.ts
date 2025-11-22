import { forwardRef, Module } from '@nestjs/common';

import { AuthController } from './controllers/auth.controller';

import { AccessJwtModule } from './jwt';

@Module({
  imports: [AccessJwtModule],
  controllers: [AuthController],
  providers: [],
  exports: [AccessJwtModule],
})
export class AuthModule {}
