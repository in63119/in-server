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
import { Contract } from 'ethers';
import { exceptions } from '../../../common/exception/exceptions';
import { EthersService } from 'src/modules/web3/services/ethers.service';
import { CONTRACT_NAME } from 'src/common/constant/enum';
import {
  ContractErrorLike,
  NormalizedPasskey,
  StoredPasskey,
} from 'src/common/constant/type';
import { decrypt } from 'src/common/util/crypto.utils';
import { AccessJwtService } from '../jwt/access/access-jwt.service';

@Injectable()
export class AuthService {
  private readonly logger = new Logger(AuthService.name);
  private readonly env: string;
  private readonly authHash: string;

  constructor(
    private readonly configService: ConfigService,
    private readonly ethersService: EthersService,
    private readonly accessJwtService: AccessJwtService,
  ) {
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

  responseAuthenticationOption = async (email: string) => {
    try {
      const rpId = this.RpID();
      const address = await this.ethersService.wallet(email).getAddress();

      const passkeys = (await this.getPasskeys(address)) as NormalizedPasskey[];
      if (passkeys.length === 0) {
        throw exceptions.Auth.NO_PASSKEY;
      }

      const credentialIds = passkeys
        .map((pk) => pk.credential.idBase64Url)
        .filter((value): value is string => Boolean(value));
      if (credentialIds.length === 0) {
        throw exceptions.Auth.NO_PASSKEY;
      }
      const options = await this.generateAuthenticaterOptions(rpId, passkeys);

      const jwt = this.accessJwtService.generate(
        email,
        options.challenge,
        credentialIds,
      );
      return { options, jwt };
    } catch (error) {
      this.logger.error('[AuthService]responseAuthenticationOption', error);
      throw exceptions.Auth.REGISTRATION_OPTIONS_ERROR;
    }
  };

  getPasskeys = async (address: string) => {
    try {
      const contract = this.ethersService.getContract(
        CONTRACT_NAME.AUTHSTORAGE,
      );
      const relayer = await this.ethersService.getReadyRelayerWallet();
      const authStorage = contract.connect(relayer) as Contract & {
        getPasskeys: (
          address: string,
        ) => Promise<[bigint, bigint, string, string][]>;
      };

      const rawPasskeys = await authStorage.getPasskeys(address);
      const passkeys = rawPasskeys
        .filter(
          ([, , credentialId, encrypted]: [bigint, bigint, string, string]) =>
            credentialId.length > 0 && encrypted.length > 0,
        )
        .map(([, , , encrypted]: [bigint, bigint, string, string]) =>
          this.reviveBuffers(JSON.parse(decrypt(encrypted, this.AuthHash()))),
        );

      return passkeys;
    } catch (error: unknown) {
      const reason = this.extractContractErrorReason(error);

      if (reason === 'AuthStorage: user not registered') {
        throw exceptions.User.USER_NOT_FOUND;
      }
    }
  };

  reviveBuffers = (passkey: StoredPasskey): NormalizedPasskey => {
    const { credential } = passkey;
    if (!credential || typeof credential.id !== 'string') {
      throw new Error('Invalid passkey credential format.');
    }

    const {
      id,
      publicKey,
      transports: storedTransports,
      counter,
      ...rest
    } = credential;
    const idBase64 = id;
    const idBuffer = this.decodeBase64(idBase64);
    const transportsSource = storedTransports ?? [];
    const transports = transportsSource.filter(
      (transport): transport is AuthenticatorTransportFuture =>
        typeof transport === 'string',
    );
    const publicKeyBuffer = this.decodeBase64(publicKey);
    const publicKeyArrayBuffer = new ArrayBuffer(publicKeyBuffer.length);
    const publicKeyUint8: Uint8Array<ArrayBuffer> = new Uint8Array(
      publicKeyArrayBuffer,
    );
    publicKeyUint8.set(publicKeyBuffer);

    return {
      ...passkey,
      credential: {
        ...rest,
        id: idBase64,
        idBuffer,
        idBase64,
        idBase64Url: this.bufferToBase64Url(idBuffer),
        publicKey: publicKeyUint8,
        publicKeyBuffer,
        counter,
        transports,
      },
      attestationObject:
        typeof passkey.attestationObject === 'string'
          ? this.decodeBase64(passkey.attestationObject)
          : undefined,
    };
  };

  extractContractErrorReason = (error: unknown): string => {
    if (typeof error === 'object' && error !== null) {
      const {
        reason,
        shortMessage,
        error: nested,
      } = error as ContractErrorLike;
      return (
        reason ?? shortMessage ?? nested?.message ?? 'Unknown contract error'
      );
    }
    return 'Unknown contract error';
  };

  decodeBase64 = (value: string) => {
    const normalized = value.replace(/-/g, '+').replace(/_/g, '/');
    const padding = '='.repeat((4 - (normalized.length % 4)) % 4);
    return Buffer.from(normalized + padding, 'base64');
  };

  bufferToBase64Url = (buffer: Buffer) =>
    buffer
      .toString('base64')
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '');

  generateAuthenticaterOptions = async (
    rpID: string,
    passkeys: NormalizedPasskey[],
  ) => {
    const allowCredentials = passkeys.map((pk) => ({
      id: pk.credential.idBase64Url,
      type: 'public-key' as const,
      transports: pk.credential.transports,
    }));
    const options = await generateAuthenticationOptions({
      rpID,
      allowCredentials,
      userVerification: 'required',
      timeout: 60000,
    });

    const credentialIds = passkeys
      .map((pk) => pk.credential.idBase64Url)
      .filter((value): value is string => Boolean(value));

    if (credentialIds.length === 0) {
      throw exceptions.Auth.NO_PASSKEY;
    }

    return options;
  };
}
