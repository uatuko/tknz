import { env } from '$env/dynamic/private';

import { redirect } from '@sveltejs/kit'

import { home_path, sign_in_path } from '$lib/server/consts';
import { check } from '$lib/server/sessions'

export async function load({ cookies, url }) {
	if (await check(cookies)) {
		return redirect(302, home_path);
	}

	const code = url.searchParams.get('code');
	if (!code) {
		return redirect(302, authzEndpoint(url))
	}

	// todo: exchange code for token
	// todo: start session

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
