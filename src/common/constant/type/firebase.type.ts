import { RELAYER_STATUS } from '../enum';

export type FirebaseRelayer = {
  address?: string;
  status?: RELAYER_STATUS;
};
