import { end } from '$lib/server/sessions';

export function load({ cookies }) {
	end(cookies);
}
