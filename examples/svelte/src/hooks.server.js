import { redirect } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';

import { public_paths, sign_in_path } from '$lib/server/consts';
import { check } from '$lib/server/sessions';

/** @type {import('@sveltejs/kit').Handle} */
async function authHandle({ event, resolve }) {
	if (public_paths.includes(event.url.pathname)) {
		return resolve(event);
	}

	if (!(await check(event.cookies))) {
		return redirect(302, sign_in_path);
	}

	return resolve(event);
}

export const handle = sequence(authHandle);
