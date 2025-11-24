import { forwardRef, Inject, Injectable } from '@nestjs/common';
import {
  toUtf8Bytes,
  Wallet,
  keccak256,
  JsonRpcProvider,
  Contract,
  FeeData,
  encodeBytes32String,
  Interface,
} from 'ethers';
import { decrypt } from 'src/common/util/crypto.utils';
import { ConfigService } from '@nestjs/config';
import { getAbis } from 'src/common/abis';
import { AuthService } from 'src/modules/auth/services/auth.service';
import { exceptions } from '../../../common/exception/exceptions';
import { CONTRACT_NAME, RELAYER_STATUS } from 'src/common/constant/enum';
import { FirebaseService } from '../../db/services/firebase.service';
import { FirebaseRelayer } from 'src/common/constant/type';

@Injectable()
export class EthersService {
  private readonly authStorage: Contract;
  private readonly postStorage: Contract;
  private readonly postForwarder: Contract;
  private readonly visitorStorage: Contract;
  private readonly youtubeStorage: Contract;
  private readonly subscriberStorage: Contract;
  private readonly relayerManager: Contract;

  private readonly provider = new JsonRpcProvider(
    'https://public-en-kairos.node.kaia.io',
  );

  private readonly ownerPK: string;
  private readonly relayerPK: string;
  private readonly relayer2PK: string;
  private readonly relayer3PK: string;

  constructor(
    private readonly configService: ConfigService,
    @Inject(forwardRef(() => AuthService))
    private readonly authService: AuthService,
    private readonly firebaseService: FirebaseService,
  ) {
    const abis = getAbis();

    this.authStorage = new Contract(
      abis.AuthStorage.address,
      abis.AuthStorage.abi,
    );
    this.postStorage = new Contract(
      abis.PostStorage.address,
      abis.PostStorage.abi,
    );
    this.postForwarder = new Contract(
      abis.PostForwarder.address,
      abis.PostForwarder.abi,
    );
    this.visitorStorage = new Contract(
      abis.VisitorStorage.address,
      abis.VisitorStorage.abi,
    );
    this.youtubeStorage = new Contract(
      abis.YoutubeStorage.address,
      abis.YoutubeStorage.abi,
    );
    this.subscriberStorage = new Contract(
      abis.SubscriberStorage.address,
      abis.SubscriberStorage.abi,
    );
    this.relayerManager = new Contract(
      abis.RelayerManager.address,
      abis.RelayerManager.abi,
    );

    this.ownerPK =
      this.configService.get<string>('blockchain.privateKey.owner') || '';
    this.relayerPK =
      this.configService.get<string>('blockchain.privateKey.relayer') || '';
    this.relayer2PK =
      this.configService.get<string>('blockchain.privateKey.relayer2') || '';
    this.relayer3PK =
      this.configService.get<string>('blockchain.privateKey.relayer3') || '';
  }

  wallet = (email: string) => {
    const salt = this.authService.AuthHash();
    if (salt.length === 0) {
      throw exceptions.System.INVALID_AUTH_HASH;
    }
    const input = toUtf8Bytes(email + salt);
    const privateKeyBytes = keccak256(input);
    const wallet = new Wallet(privateKeyBytes);

    return wallet;
  };

  accounts = () => {
    const accountsPKs = {
      owner: this.ownerPK,
      relayer: this.relayerPK,
      relayer2: this.relayer2PK,
      relayer3: this.relayer3PK,
    };
    if (
      accountsPKs.owner.length === 0 ||
      accountsPKs.relayer.length === 0 ||
      accountsPKs.relayer2.length === 0 ||
      accountsPKs.relayer3.length === 0
    ) {
      throw exceptions.System.INVALID_PRIVATE_KEY;
    }

    const salt = this.authService.AuthHash();
    if (salt.length === 0) {
      throw exceptions.System.INVALID_AUTH_HASH;
    }

    return {
      owner: new Wallet(decrypt(accountsPKs.owner, salt), this.provider),
      relayer: new Wallet(decrypt(accountsPKs.relayer, salt), this.provider),
      relayer2: new Wallet(decrypt(accountsPKs.relayer2, salt), this.provider),
      relayer3: new Wallet(decrypt(accountsPKs.relayer3, salt), this.provider),
    };
  };

  getContract = (contract: CONTRACT_NAME) => {
    switch (contract) {
      case CONTRACT_NAME.POSTSTORAGE:
        return this.postStorage;

      case CONTRACT_NAME.AUTHSTORAGE:
        return this.authStorage;

      case CONTRACT_NAME.POSTFORWARDER:
        return this.postForwarder;

      case CONTRACT_NAME.RELAYERMANAGER:
        return this.relayerManager;

      case CONTRACT_NAME.VISITORSTORAGE:
        return this.visitorStorage;

      case CONTRACT_NAME.YOUTUBESTORAGE:
        return this.youtubeStorage;

      case CONTRACT_NAME.SUBSCRIBERSTORAGE:
        return this.subscriberStorage;
      default:
        throw exceptions.Blockchain.CONTRACT_NOT_FOUND(contract);
    }
  };

  getReadyRelayerWallet = async () => {
    const { relayer, relayer2, relayer3 } = this.accounts();
    const firebaseRelayers =
      (await this.firebaseService.read<Record<string, FirebaseRelayer>>(
        'relayers',
      )) ?? {};

    const availableRelayers = [relayer, relayer2, relayer3];

    for (const walletCandidate of availableRelayers) {
      const walletAddress = walletCandidate.address.toLowerCase();
      const isReady = Object.values(firebaseRelayers).some(
        (entry) =>
          entry?.status === RELAYER_STATUS.Ready &&
          entry.address?.toLowerCase() === walletAddress,
      );
      if (isReady) {
        return walletCandidate;
      }
    }

    throw exceptions.Blockchain.NO_AVAILABLE_RELAYER;
  };
}
