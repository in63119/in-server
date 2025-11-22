import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { App } from 'firebase-admin/app';
import admin, { ServiceAccount } from 'firebase-admin';
import { Database } from 'firebase-admin/database';
import { Reference } from 'firebase-admin/database';

@Injectable()
export class FirebaseService {
  private readonly app: App;
  private readonly db: Database;

  constructor(private readonly configService: ConfigService) {
    const databaseURL = this.configService.get<string>('firebase.databaseURL');
    const projectId = this.configService.get<string>('firebase.project_id');
    const clientEmail = this.configService.get<string>('firebase.client_email');
    const privateKeyRaw = this.configService.get<string>(
      'firebase.private_key',
    );

    if (!databaseURL || !projectId || !clientEmail || !privateKeyRaw) {
      throw new Error('Firebase configuration is missing');
    }

    const serviceAccount: ServiceAccount = {
      projectId,
      clientEmail,
      privateKey: privateKeyRaw.replace(/\\n/g, '\n'),
    };
    admin.initializeApp({
      credential: admin.credential.cert(serviceAccount),
      databaseURL,
    });

    this.app = admin.app();
    this.db = admin.database(this.app);
  }

  getRef = (path: string): Reference => {
    if (!path || typeof path !== 'string') {
      throw new Error('Firebase path must be a non-empty string');
    }

    const normalizedPath = path.replace(/^\/+/, '');
    return this.db.ref(normalizedPath);
  };

  read = async <T = unknown>(path: string): Promise<T | null> => {
    const snapshot = await this.getRef(path).get();
    if (!snapshot.exists()) {
      return null;
    }

    return snapshot.val() as T;
  };

  write = async <T>(path: string, value: T): Promise<void> => {
    await this.getRef(path).set(value);
  };
}
