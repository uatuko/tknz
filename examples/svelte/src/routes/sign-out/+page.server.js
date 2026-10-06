import { end } from '#lib/server/sessions.js';

export function load({ cookies }) {
	end(cookies);
}
