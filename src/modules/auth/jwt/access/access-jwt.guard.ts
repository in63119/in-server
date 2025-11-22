import { Injectable, CanActivate, ExecutionContext } from '@nestjs/common';
import { Request } from 'express';
import { AccessJwtService } from './access-jwt.service';
import { exceptions } from '../../../../common/exception/exceptions';

@Injectable()
export class AccessJwtGuard implements CanActivate {
  constructor(private readonly jwtService: AccessJwtService) {}

  async canActivate(ctx: ExecutionContext): Promise<boolean> {
    const req = ctx.switchToHttp().getRequest<Request>();
    const auth = req.headers.authorization;
    if (!auth) throw exceptions.Auth.INVALID_AUTHORIZATION;

    const [type, token] = auth.split(' ');
    if (type !== 'Bearer' || !token) {
      throw exceptions.Auth.INVALID_AUTHORIZATION_FORMAT;
    }

    let payload;

    try {
      payload = this.jwtService.verify(token);
    } catch {
      throw exceptions.Auth.INVALID_AUTHORIZATION;
    }

    (req as any).user = payload;
    return true;
  }
}
