import { Module } from '@nestjs/common';
import { JwtModule } from '@nestjs/jwt';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { AccessJwtService } from './access-jwt.service';
import { AccessJwtGuard } from './access-jwt.guard';

@Module({
  imports: [
    JwtModule.registerAsync({
      imports: [ConfigModule],
      inject: [ConfigService],
      useFactory: (cfg: ConfigService) => ({
        secret: cfg.get<string>('jwt.accessSecret'),
        signOptions: { expiresIn: '7d' },
      }),
    }),
  ],
  providers: [AccessJwtService, AccessJwtGuard],
  exports: [AccessJwtService, AccessJwtGuard],
})
export class AccessJwtModule {}
