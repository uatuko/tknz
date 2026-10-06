import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import { tknz_addr } from '$app/env/private';

import authnProto from './proto/authn.json';

/**
 * @import { ProtoGrpcType } from '#lib/server/proto/authn.js'
 * @import { AuthnClient } from '#lib/server/proto/tknz/v1/Authn.js'
 *
 * @import {INamespace} from 'protobufjs'
 */

/** @type { AuthnClient | undefined } */
let _client;

export const cookie_same_site_lax = 'lax';

export const session_cookie_max_age = 60 * 60 * 24; // 1 day
export const session_cookie_name = 'id';

/**
 * Check a session is valid.
 *
 * @param {import('@sveltejs/kit').Cookies} cookies - Cookies interface
 */
export async function check(cookies) {
	const token = cookies.get(session_cookie_name);
	if (token === undefined) {
		return false;
	}

	return new Promise((resolve, reject) => {
		client().Check({ token }, (err, resp) => {
			if (err) {
				reject(err);
			}

			if (resp === undefined) {
				return reject(new Error('empty response from server'));
			}

			resolve(resp.ok);
		});
	});
}

/**
 * End a session.
 *
 * @param {import('@sveltejs/kit').Cookies} cookies - Cookies interface
 */
export function end(cookies) {
	cookies.set(session_cookie_name, '', {
		httpOnly: true,
		path: '/',
		secure: import.meta.env.PROD,
		sameSite: cookie_same_site_lax,
		maxAge: 0,
	});
}

/**
 * Start a new session.
 *
 * @param {import('@sveltejs/kit').Cookies} cookies - Cookies interface
 * @param {string} token - Access token
 */
export function start(cookies, token) {
	cookies.set(session_cookie_name, token, {
		httpOnly: true,
		path: '/',
		secure: import.meta.env.PROD,
		sameSite: cookie_same_site_lax,
		maxAge: session_cookie_max_age,
	});
}

function client() {
	if (_client) {
		return _client;
	}

	const defs = /** @type {ProtoGrpcType} */ (
		/** @type {unknown} */ (
			grpc.loadPackageDefinition(protoLoader.fromJSON(/** @type {INamespace} */ (authnProto)))
		)
	);

	let opts;
	if (tknz_addr.startsWith('localhost:')) {
		opts = grpc.credentials.createInsecure();
	} else {
		opts = grpc.credentials.createSsl();
	}

	_client = new defs.tknz.v1.Authn(tknz_addr, opts);

	return _client;
}
