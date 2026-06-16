import { goto } from '$app/navigation';

export function load({ url }) {
	const state = url.searchParams.get('state');
	const loginHint = url.searchParams.get('login_hint');
	const errorCode = url.searchParams.get('error_code');

	const backParams = new URLSearchParams();
	if (state) backParams.append('state', state);
	if (loginHint) backParams.append('login_hint', loginHint);
	const backUri = `/sign-in?${backParams.toString()}`;

	if (!loginHint) goto(backUri);

	let error = null;
	switch (errorCode) {
		case null:
			break;

		case 'invalid_credentials':
			error = 'Invalid username or password';
			break;

		default:
			error = 'There was an issue processing your request';
			break;
	}

	return {
		error,
		loginHint,
		state,
		backUri,
	};
}
