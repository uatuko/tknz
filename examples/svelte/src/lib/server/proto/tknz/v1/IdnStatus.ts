// Original file: ../../proto/tknz/v1/authn.proto

export const IdnStatus = {
  idn_pending: 0,
  idn_active: 1,
  idn_suspended: -1,
  idn_deleted: -2,
} as const;

export type IdnStatus =
  | 'idn_pending'
  | 0
  | 'idn_active'
  | 1
  | 'idn_suspended'
  | -1
  | 'idn_deleted'
  | -2

export type IdnStatus__Output = typeof IdnStatus[keyof typeof IdnStatus]
