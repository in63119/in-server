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
  REGISTRATION_OPTIONS_ERROR: createException(
    'REGISTRATION_OPTIONS_ERROR',
    500,
  ),
  AUTHENTICATION_OPTIONS_ERROR: createException(
    'AUTHENTICATION_OPTIONS_ERROR',
    500,
  ),
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

const SystemException = {
  INVALID_AUTH_HASH: createException('INVALID_AUTH_HASH', 500),
  INVALID_PRIVATE_KEY: createException('INVALID_PRIVATE_KEY', 500),
};

const BlockchainException = {
  CONTRACT_NOT_FOUND: (contractName: string) =>
    createException(
      {
        message: 'CONTRACT_NOT_FOUND',
        contractName,
      },
      500,
    ),
  NO_AVAILABLE_RELAYER: createException('NO_AVAILABLE_RELAYER', 500),
};

const UserException = {
  USER_NOT_FOUND: createException('USER_NOT_FOUND', 404),
};

export const exceptions = {
  Auth: AuthException,
  System: SystemException,
  Blockchain: BlockchainException,
  User: UserException,

  createBadRequestException: (message: string) => createException(message, 400),
};
