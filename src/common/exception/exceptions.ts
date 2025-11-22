import { HttpException } from '@nestjs/common';

function createException(
  message: string | object,
  statusCode: number,
  description?: string,
): HttpException {
  return new HttpException(message, statusCode, { description });
}

const AuthException = {
  INVALID_ORIGIN: createException('INVALID_ORIGIN', 400),
  INVALID_GRANT: createException('INVALID_GRANT', 400),
  FAILED_GENERATE_OPTIONS: createException('FAILED_GENERATE_OPTIONS', 500),
  INVALID_REGISTRATION_TOKEN: createException(
    'INVALID_REGISTRATION_TOKEN',
    400,
  ),
  INVALID_AUTHORIZATION: createException('INVALID_AUTHORIZATION', 401),
  INVALID_AUTHORIZATION_FORMAT: createException('INVALID_AUTHORIZATION', 401),
  FAILED_VERIFY_CREDENTIAL: createException('FAILED_VERIFY_CREDENTIAL', 400),
  INVALID_CHALLENGE: createException('INVALID_CHALLENGE', 400),
  ALREADY_REGISTERED: (existingName: string) =>
    createException(
      {
        message: 'ALREADY_REGISTERED',
        existingName,
      },
      400,
    ),
  NO_PASSKEY: createException('NO_PASSKEY', 400),
};

export const exceptions = {
  Auth: AuthException,

  createBadRequestException: (message: string) => createException(message, 400),
};
