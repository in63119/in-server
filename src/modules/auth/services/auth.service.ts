import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import {
  generateRegistrationOptions,
  verifyRegistrationResponse,
  generateAuthenticationOptions,
  verifyAuthenticationResponse,
  VerifiedRegistrationResponse,
  AuthenticatorTransportFuture,
  VerifiedAuthenticationResponse,
} from '@simplewebauthn/server';

import { exceptions } from '../../../common/exception/exceptions';

@Injectable()
export class AuthService {
  private readonly logger = new Logger(AuthService.name);
  private readonly env: string;
  private readonly authHash: string;

  constructor(private readonly configService: ConfigService) {
    this.env = this.configService.get<string>('ENV') || '';
    this.authHash = this.configService.get<string>('auth.hash') || '';
  }

  RpID = () => {
    let result: string;

    if (this.env === 'development') {
      result = 'localhost';
    } else if (this.env === 'production') {
      result = 'in-labs.xyz';
    } else {
      throw exceptions.Auth.INVALID_ORIGIN;
    }

    return result;
  };

  AuthHash = () => {
    return this.authHash;
  };
}
