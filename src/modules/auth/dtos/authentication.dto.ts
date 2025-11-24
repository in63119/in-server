import { IsString } from 'class-validator';

export class AuthenticationOptionsDto {
  @IsString()
  email: string;
}
