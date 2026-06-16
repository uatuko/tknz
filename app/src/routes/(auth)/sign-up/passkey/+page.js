import { error } from '@sveltejs/kit';

import { decode } from '$lib/b64.js';
import { http_status_bad_request } from '$lib/consts.js';

export function load({ url }) {
	const challenge = url.searchParams.get('challenge');
	const login = url.searchParams.get('login_hint');
	const rpId = url.searchParams.get('rp_id');
	const rpName = url.searchParams.get('rp_name');
	const state = url.searchParams.get('state');
	const uid = url.searchParams.get('uid');

	if (!challenge || !login || !rpId || !rpName || !state || !uid) {
		error(http_status_bad_request, 'invalid or malformed request');
	}

	return {
		challenge: decode(challenge),
		rp: {
			id: rpId,
			name: rpName,
		},
		state,
		user: {
			displayName: '',
			id: decode(challenge),
			name: login,
		},
	};
}
