import { Controller, Get, Post, Body, Req, UseGuards } from '@nestjs/common';
import { routes } from '../../../common/apis/endpoints';
import { AuthService } from '../services/auth.service';
// import { Request } from 'express';

import { AuthenticationOptionsDto } from '../dtos';

// import { RegistrationJwtGuard } from '../jwt';

@Controller(routes.auth.root)
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Post(routes.auth.authentication.option)
  authenticationOptions(@Body() { email }: AuthenticationOptionsDto) {
    return this.authService.responseAuthenticationOption(email);
  }
}
