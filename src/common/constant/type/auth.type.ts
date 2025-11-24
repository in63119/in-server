import {
  AuthenticatorTransportFuture,
  RegistrationResponseJSON,
} from '@simplewebauthn/server';
import { Device, OS } from '../enum';

export interface RequestWebauthnOptions {
  email: string;
  device?: Device;
  os?: OS;
  allowMultipleDevices?: boolean;
}

export type NormalizedPasskey = {
  credential: {
    id: string;
    idBuffer: Buffer;
    idBase64: string;
    idBase64Url: string;
    publicKey: Uint8Array<ArrayBuffer>;
    publicKeyBuffer: Buffer;
    counter: number;
    transports: AuthenticatorTransportFuture[];
    [key: string]: unknown;
  };
  attestationObject?: Buffer;
  [key: string]: unknown;
};

export type PasskeySummary = {
  address: string;
  credentialId: string;
  transports: AuthenticatorTransportFuture[];
  counter: number;
  deviceType?: string | null;
  adminDeviceType?: string | null;
  osLabel?: string | null;
  backedUp?: boolean | null;
};

export type RegistrationRequest = {
  credential: RegistrationResponseJSON;
  device: Device;
  os?: OS;
  allowMultipleDevices: boolean;
};

export type DeviceInfoLike = {
  deviceType?: string | null;
  os?: string | null;
  userAgent?: string | null;
};

export type StoredPasskey = {
  credential: {
    id: string;
    publicKey: string;
    counter: number;
    transports?: AuthenticatorTransportFuture[] | string[];
  };
  attestationObject?: string;
  [key: string]: unknown;
};
