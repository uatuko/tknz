import { cookie_same_site_lax } from '$lib/server/consts';

export const session_cookie_max_age = 60 * 60 * 24; // 1 day
export const session_cookie_name = 'id';

/**
 * Check a session is valid.
 *
 * @param {import('@sveltejs/kit').Cookies} cookies - Cookies interface
 */
export async function check(cookies) {
	return false;
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
