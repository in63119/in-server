import { Controller, Get, Post, Body, Req, UseGuards } from '@nestjs/common';
import { routes } from '../../../common/apis/endpoints';
// import { Request } from 'express';

// import { AuthService } from '@modules/auth/services/auth.service';

// import {
//   OptionsDto,
//   RegistrationCredentialDto,
//   AuthenticationCredentialDto,
//   SignUpDto,
//   SignInDto,
// } from '@modules/auth/dtos';

// import { RegistrationJwtGuard } from '../jwt';

@Controller(routes.auth.root)
export class AuthController {
  //   constructor(private readonly authService: AuthService) {}
  //   @Post(routes.auth.signin)
  //   signIn(@Body() { kakaoId }: SignInDto) {
  //     return this.authService.signIn(kakaoId);
  //   }
  //   @Post(routes.auth.signup)
  //   signUp(@Body() signUpDto: SignUpDto) {
  //     return this.authService.signUp(signUpDto);
  //   }
  //   @Post(routes.auth.options)
  //   options(@Body() optionsDto: OptionsDto, @Req() req: Request) {
  //     const origin = req.headers.origin;
  //     return this.authService.generateOptions(optionsDto, origin);
  //   }
  //   @Post(routes.auth.verify)
  //   verify(@Body() credential: AuthenticationCredentialDto) {
  //     return this.authService.verifyOptions(credential);
  //   }
  //   @Post(routes.auth.registerOptions)
  //   registerOptions(@Body() optionsDto: OptionsDto, @Req() req: Request) {
  //     const origin = req.headers.origin;
  //     return this.authService.generateRegisterOptions(optionsDto, origin);
  //   }
  //   @UseGuards(RegistrationJwtGuard)
  //   @Post(routes.auth.registerVerify)
  //   registerVerify(
  //     @Req() req: Request,
  //     @Body() credential: RegistrationCredentialDto,
  //   ) {
  //     const { challengeUserId, userName, kakaoId } = (req as any).user;
  //     return this.authService.verifyRegisterCredential(credential, {
  //       challengeUserId,
  //       userName,
  //       kakaoId,
  //     });
  //   }
}
