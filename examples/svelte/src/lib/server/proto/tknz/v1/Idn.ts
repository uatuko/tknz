// Original file: ../../proto/tknz/v1/authn.proto

import type { IdnStatus as _tknz_v1_IdnStatus, IdnStatus__Output as _tknz_v1_IdnStatus__Output } from '../../tknz/v1/IdnStatus';

export interface Idn {
  'id'?: (string);
  'status'?: (_tknz_v1_IdnStatus);
  'spaceId'?: (string);
  'login'?: (string);
  'email'?: (string);
  'name'?: (string);
  'picture'?: (string);
  '_email'?: "email";
  '_name'?: "name";
  '_picture'?: "picture";
}

export interface Idn__Output {
  'id': (string);
  'status': (_tknz_v1_IdnStatus__Output);
  'spaceId': (string);
  'login': (string);
  'email'?: (string);
  'name'?: (string);
  'picture'?: (string);
  '_email'?: "email";
  '_name'?: "name";
  '_picture'?: "picture";
}
