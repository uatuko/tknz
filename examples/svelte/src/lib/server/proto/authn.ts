import type * as grpc from '@grpc/grpc-js';
import type { EnumTypeDefinition, MessageTypeDefinition } from '@grpc/proto-loader';

import type { AuthnClient as _tknz_v1_AuthnClient, AuthnDefinition as _tknz_v1_AuthnDefinition } from './tknz/v1/Authn';
import type { AuthnCheckRequest as _tknz_v1_AuthnCheckRequest, AuthnCheckRequest__Output as _tknz_v1_AuthnCheckRequest__Output } from './tknz/v1/AuthnCheckRequest';
import type { AuthnCheckResponse as _tknz_v1_AuthnCheckResponse, AuthnCheckResponse__Output as _tknz_v1_AuthnCheckResponse__Output } from './tknz/v1/AuthnCheckResponse';
import type { Idn as _tknz_v1_Idn, Idn__Output as _tknz_v1_Idn__Output } from './tknz/v1/Idn';

type SubtypeConstructor<Constructor extends new (...args: any) => any, Subtype> = {
  new(...args: ConstructorParameters<Constructor>): Subtype;
};

export interface ProtoGrpcType {
  tknz: {
    v1: {
      Authn: SubtypeConstructor<typeof grpc.Client, _tknz_v1_AuthnClient> & { service: _tknz_v1_AuthnDefinition }
      AuthnCheckRequest: MessageTypeDefinition<_tknz_v1_AuthnCheckRequest, _tknz_v1_AuthnCheckRequest__Output>
      AuthnCheckResponse: MessageTypeDefinition<_tknz_v1_AuthnCheckResponse, _tknz_v1_AuthnCheckResponse__Output>
      Idn: MessageTypeDefinition<_tknz_v1_Idn, _tknz_v1_Idn__Output>
      IdnStatus: EnumTypeDefinition
    }
  }
}

