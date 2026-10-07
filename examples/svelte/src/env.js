import { defineEnvVars } from '@sveltejs/kit/env';

export const variables = defineEnvVars({
	token_endpoint: { schema: (input) => input ?? '' },
	client_id: { schema: (input) => input ?? '' },
	private_key: { schema: (input) => input ?? '' },
	authorization_endpoint: { schema: (input) => input ?? '' },
	tknz_addr: { schema: (input) => input ?? '' },
});
