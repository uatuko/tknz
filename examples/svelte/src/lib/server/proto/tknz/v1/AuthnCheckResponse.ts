// Original file: ../../proto/tknz/v1/authn.proto

import type { Idn as _tknz_v1_Idn, Idn__Output as _tknz_v1_Idn__Output } from '../../tknz/v1/Idn';

export interface AuthnCheckResponse {
  'ok'?: (boolean);
  'idn'?: (_tknz_v1_Idn | null);
  '_idn'?: "idn";
}

export interface AuthnCheckResponse__Output {
  'ok': (boolean);
  'idn'?: (_tknz_v1_Idn__Output | null);
  '_idn'?: "idn";
}
