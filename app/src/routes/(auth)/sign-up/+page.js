export function load({ url }) {
	const state = url.searchParams.get('state');
	const providerId = url.searchParams.get('provider_id');

	const errorCode = url.searchParams.get('error_code');
	let error;
	switch (errorCode) {
		case null:
			break;

		case 'invalid_login':
			error = 'Invalid email address';
			break;

		default:
			error = 'There was an issue processing your request';
			break;
	}

	return {
		error,
		providerId,
		state,
	};
}
