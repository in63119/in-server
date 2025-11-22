import { Module } from '@nestjs/common';
import { FirebaseService } from './services/firebase.service';

@Module({
  imports: [],
  providers: [FirebaseService],
  exports: [FirebaseService],
})
export class DBModule {}
