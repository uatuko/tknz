import { error } from '@sveltejs/kit';

import {
	auth_providers_path,
	http_status_ok,
	http_status_bad_request,
	google_authorization_endpoint,
	provider_slug_google_oauth,
} from '$lib/consts';

/**
 * @typedef {{id: string, slug: string}} Provider
 * @typedef { Provider & {authorization_endpoint: string, client_id: string, redirect_uri: string }} OidcProvider
 *
 * @typedef {{oidc: Array<OidcProvider>, sign_up?: Provider, use_login: boolean}} ProvidersResponse
 */

/** @type {import('./$types').PageLoad} */
export async function load({ url }) {
	const providersUri = `${url.origin}${auth_providers_path}`;
	const state = url.searchParams.get('state') ?? '';
	const loginHint = url.searchParams.get('login_hint') ?? '';

	async function fetchProviders() {
		if (!state) {
			error(http_status_bad_request, 'invalid or malformed request');
		}

		const resp = await fetch(`${providersUri}?${new URLSearchParams({ state })}`);

		if (resp.status !== http_status_ok) {
			error(resp.status, 'unsuccessful response from server');
		}

		/** @type {ProvidersResponse} */
		const providers = await resp.json();

		for (const p of providers.oidc) {
			switch (p.slug) {
				case provider_slug_google_oauth: {
					p.authorization_endpoint = google_authorization_endpoint;

					break;
				}
			}
		}

		return providers;
	}

	return {
		loginHint,
		state,
		providers: fetchProviders(),
	};
}
