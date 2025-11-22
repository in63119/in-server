import { HttpException, HttpStatus } from '@nestjs/common';

export class ExceptionBase extends HttpException {
  constructor(
    className: string,
    objectOrError?: string | object | any,
    description?: string,
    status?: number,
  ) {
    if (
      objectOrError &&
      typeof objectOrError === 'object' &&
      'errCode' in objectOrError &&
      'errorMessage' in objectOrError
    ) {
      objectOrError.message = `[${className}:${objectOrError.errCode}]: ${objectOrError.errorMessage}`;
    }

    super(
      HttpException.createBody(
        objectOrError ?? 'Unknown error',
        description ?? 'Bad Request',
        status ?? HttpStatus.BAD_REQUEST,
      ),
      status ?? HttpStatus.BAD_REQUEST,
    );
  }
}
