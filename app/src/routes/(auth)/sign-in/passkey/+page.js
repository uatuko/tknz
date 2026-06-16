import { error } from '@sveltejs/kit';
import { decode as cborDecode } from 'cbor2';

import { http_status_bad_request } from '$lib/consts.js';
import { decode } from '$lib/b64.js';

export function load({ url }) {
	const state = url.searchParams.get('state');
	const data = url.searchParams.get('data');

	if (!state || !data) {
		error(http_status_bad_request, 'invalid or malformed request');
	}

	/** @type {[Uint8Array<ArrayBuffer>, Array<Uint8Array<ArrayBuffer>>]} */
	const [challenge, credentialIds] = cborDecode(decode(data));

	return {
		challenge,
		credentialIds,
		state,
	};
}
