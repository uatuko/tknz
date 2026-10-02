// Original file: ../../proto/tknz/v1/authn.proto

import type * as grpc from '@grpc/grpc-js'
import type { MethodDefinition } from '@grpc/proto-loader'
import type { AuthnCheckRequest as _tknz_v1_AuthnCheckRequest, AuthnCheckRequest__Output as _tknz_v1_AuthnCheckRequest__Output } from '../../tknz/v1/AuthnCheckRequest';
import type { AuthnCheckResponse as _tknz_v1_AuthnCheckResponse, AuthnCheckResponse__Output as _tknz_v1_AuthnCheckResponse__Output } from '../../tknz/v1/AuthnCheckResponse';

export interface AuthnClient extends grpc.Client {
  Check(argument: _tknz_v1_AuthnCheckRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_tknz_v1_AuthnCheckResponse__Output>): grpc.ClientUnaryCall;
  Check(argument: _tknz_v1_AuthnCheckRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_tknz_v1_AuthnCheckResponse__Output>): grpc.ClientUnaryCall;
  Check(argument: _tknz_v1_AuthnCheckRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_tknz_v1_AuthnCheckResponse__Output>): grpc.ClientUnaryCall;
  Check(argument: _tknz_v1_AuthnCheckRequest, callback: grpc.requestCallback<_tknz_v1_AuthnCheckResponse__Output>): grpc.ClientUnaryCall;
  check(argument: _tknz_v1_AuthnCheckRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_tknz_v1_AuthnCheckResponse__Output>): grpc.ClientUnaryCall;
  check(argument: _tknz_v1_AuthnCheckRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_tknz_v1_AuthnCheckResponse__Output>): grpc.ClientUnaryCall;
  check(argument: _tknz_v1_AuthnCheckRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_tknz_v1_AuthnCheckResponse__Output>): grpc.ClientUnaryCall;
  check(argument: _tknz_v1_AuthnCheckRequest, callback: grpc.requestCallback<_tknz_v1_AuthnCheckResponse__Output>): grpc.ClientUnaryCall;
  
}

export interface AuthnHandlers extends grpc.UntypedServiceImplementation {
  Check: grpc.handleUnaryCall<_tknz_v1_AuthnCheckRequest__Output, _tknz_v1_AuthnCheckResponse>;
  
}

export interface AuthnDefinition extends grpc.ServiceDefinition {
  Check: MethodDefinition<_tknz_v1_AuthnCheckRequest, _tknz_v1_AuthnCheckResponse, _tknz_v1_AuthnCheckRequest__Output, _tknz_v1_AuthnCheckResponse__Output>
}
