import { env } from '$env/dynamic/private';

import { redirect } from '@sveltejs/kit';

import { home_path, sign_in_path } from '$lib/consts';
import { check, start } from '$lib/server/sessions';

import * as jose from 'jose';

/**
 * @typedef {{access_token: string, id_token: string, token_type: string}} TokenResponse
 */

export async function load({ cookies, fetch, url }) {
	if (await check(cookies)) {
		return redirect(302, home_path);
	}

	const code = url.searchParams.get('code');
	if (!code) {
		return redirect(302, authzEndpoint(url));
	}

	// Exchange authorisation code for tokens (using JWT for client authentication)
	const jwt = await new jose.SignJWT()
		.setProtectedHeader({ alg: 'ES256', typ: 'JWT' })
		.setAudience(env.token_endpoint)
		.setIssuer(env.client_id)
		.setSubject(env.client_id)
		.setIssuedAt()
		.setExpirationTime('5m')
		.sign(await jose.importPKCS8(env.private_key, 'ES256'));

	const resp = await fetch(env.token_endpoint, {
		method: 'post',
		headers: {
			'content-type': 'application/x-www-form-urlencoded',
		},
		// Ref: https://www.rfc-editor.org/rfc/rfc7523.html#section-2.2
		body: new URLSearchParams({
			grant_type: 'authorization_code',
			code,
			client_id: env.client_id,
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

/** @param {URL} url  */
function authzEndpoint(url) {
	const u = new URL(env.authorization_endpoint);
	u.searchParams.set('client_id', env.client_id);
	u.searchParams.set('response_type', 'code');
	u.searchParams.set('scope', 'openid email');
	u.searchParams.set('redirect_uri', `${url.origin}${sign_in_path}`);

	return u;
}
