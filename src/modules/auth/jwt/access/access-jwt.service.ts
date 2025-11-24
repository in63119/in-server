import { Injectable, UnauthorizedException } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { exceptions } from '../../../../common/exception/exceptions';

@Injectable()
export class AccessJwtService {
  constructor(private readonly jwt: JwtService) {}

  generate(email: string, challenge: string, credentialIds: string[]) {
    return this.jwt.sign({ email, challenge, credentialIds, type: 'access' });
  }

  verify(token: string) {
    const payload = this.jwt.verify<any>(token);
    if (payload.type !== 'access') {
      throw exceptions.Auth.INVALID_AUTHORIZATION;
    }
    return payload;
  }
}
