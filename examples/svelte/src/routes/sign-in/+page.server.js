import { randomBytes, timingSafeEqual } from 'node:crypto';
import { error, redirect } from '@sveltejs/kit';

import { token_endpoint, client_id, private_key, authorization_endpoint } from '$app/env/private';

import { home_path, sign_in_path } from '#lib/consts.js';
import { check, cookie_same_site_lax, start } from '#lib/server/sessions.js';

import * as jose from 'jose';

/**
 * @typedef {{access_token: string, id_token: string, token_type: string}} TokenResponse
 */

const state_cookie_max_age = 60 * 15; // 15 mins
const state_cookie_name = 'state';

export async function load({ cookies, fetch, url }) {
	if (await check(cookies)) {
		return redirect(302, home_path);
	}

	const code = url.searchParams.get('code');
	if (!code) {
		const state = randomBytes(32).toString('base64url');
		cookies.set(state_cookie_name, state, {
			httpOnly: true,
			path: sign_in_path,
			secure: import.meta.env.PROD,
			sameSite: cookie_same_site_lax,
			maxAge: state_cookie_max_age,
		});

		return redirect(302, authzEndpoint(url, state), { external: true });
	}

	// Verify state to mitigate CSRF
	const expected = cookies.get(state_cookie_name);
	cookies.delete(state_cookie_name, { path: sign_in_path }); // state is single use, clear before checking
	if (!verifyState(expected, url.searchParams.get('state'))) {
		return error(400, 'invalid state');
	}

	// Exchange authorisation code for tokens (using JWT for client authentication)
	const jwt = await new jose.SignJWT()
		.setProtectedHeader({ alg: 'ES256', typ: 'JWT' })
		.setAudience(token_endpoint)
		.setIssuer(client_id)
		.setSubject(client_id)
		.setIssuedAt()
		.setExpirationTime('5m')
		.sign(await jose.importPKCS8(private_key, 'ES256'));

	const resp = await fetch(token_endpoint, {
		method: 'post',
		headers: {
			'content-type': 'application/x-www-form-urlencoded',
		},
		// Ref: https://www.rfc-editor.org/rfc/rfc7523.html#section-2.2
		body: new URLSearchParams({
			grant_type: 'authorization_code',
			code,
			client_id,
			client_assertion_type: 'urn:ietf:params:oauth:client-assertion-type:jwt-bearer',
			client_assertion: jwt,
		}).toString(),
	});

	if (!resp.ok) {
		console.error(await resp.text());
		throw new Error('failed to exchange code for token');
	}

	/** @type {TokenResponse} */
	const tokens = await resp.json();

	// Start a new session
	start(cookies, tokens.access_token);

	redirect(302, home_path);
}

/**
 * @param {URL} url
 * @param {string} state
 */
function authzEndpoint(url, state) {
	const u = new URL(authorization_endpoint);
	u.searchParams.set('client_id', client_id);
	u.searchParams.set('response_type', 'code');
	u.searchParams.set('scope', 'openid email');
	u.searchParams.set('redirect_uri', `${url.origin}${sign_in_path}`);
	u.searchParams.set('state', state);

	return u;
}

/**
 * @param {string | undefined} expected - State value stored in the cookie
 * @param {string | null} actual - State value returned with the authorisation code
 */
function verifyState(expected, actual) {
	if (!expected || !actual) {
		return false;
	}

	const a = Buffer.from(expected);
	const b = Buffer.from(actual);

	return a.length === b.length && timingSafeEqual(a, b);
}
