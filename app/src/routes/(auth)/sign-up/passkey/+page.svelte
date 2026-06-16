<script>
	import IconFingerprint from '$lib/icons/fingerprint-outline.svelte';
	import LogoFelk from '$lib/icons/logo-felk.svelte';

	import { auth_credentials_path } from '$lib/consts.js';

	let { data } = $props();

	let busy = $state(false);
	let error = $state(false);

	async function submit(/** @type {SubmitEvent} */ e) {
		error = false;
		e.preventDefault();

		if (!(e.target instanceof HTMLFormElement)) {
			error = true;
			return;
		}

		const input = e.target.elements.namedItem('credential');
		if (!(input instanceof HTMLInputElement)) {
			error = true;
			return;
		}

		busy = true;
		const { challenge, rp, user } = data;

		/** @type {PublicKeyCredentialCreationOptions} */
		const publicKey = {
			challenge,
			rp,
			user,

			attestation: 'indirect',
			authenticatorSelection: {
				residentKey: 'discouraged',
				requireResidentKey: false,
				userVerification: 'required',
			},
			pubKeyCredParams: [
				{
					type: 'public-key',
					alg: -8, // EdDSA
				},
				{
					type: 'public-key',
					alg: -7, // ES256
				},
				{
					type: 'public-key',
					alg: -257, // RS256
				},
			],
			timeout: 300000, // 5 minutes
		};

		let cred;
		try {
			cred = await navigator.credentials.create({ publicKey });
		} catch {
			error = true;
		}

		busy = false;
		if (
			!(cred instanceof PublicKeyCredential) ||
			!(cred.response instanceof AuthenticatorAttestationResponse)
		) {
			error = true;
			return;
		}

		input.value = JSON.stringify(cred);
		e.target.submit();
	}
</script>

<div class="sm:mx-auto sm:w-full sm:max-w-sm">
	<LogoFelk
		class="mx-auto size-12 rounded-md bg-black fill-white p-3 dark:bg-white dark:fill-black"
	/>
	<h1 class="mt-6 text-center text-xl font-semibold text-gray-900 dark:text-white">
		Setup passkey
	</h1>
	<p class="mt-2 text-center text-sm/6 text-gray-600 dark:text-gray-400">
		Finish setting up your account by creating a passkey.
	</p>
</div>

<div class="mx-auto w-full max-w-sm space-y-5 sm:max-w-md">
	{#if error}
		<div
			class="max-w-xl rounded-md bg-red-50 p-4 dark:bg-red-500/15 dark:outline dark:outline-red-500/25"
		>
			<div class="flex">
				<div class="shrink-0">
					<svg
						viewBox="0 0 20 20"
						fill="currentColor"
						data-slot="icon"
						aria-hidden="true"
						class="size-5 text-red-400"
					>
						<path
							d="M10 18a8 8 0 1 0 0-16 8 8 0 0 0 0 16ZM8.28 7.22a.75.75 0 0 0-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 1 0 1.06 1.06L10 11.06l1.72 1.72a.75.75 0 1 0 1.06-1.06L11.06 10l1.72-1.72a.75.75 0 0 0-1.06-1.06L10 8.94 8.28 7.22Z"
							clip-rule="evenodd"
							fill-rule="evenodd"
						/>
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">
						Failed to create passkey
					</h3>
					<div class="mt-2 text-sm text-red-700 dark:text-red-200/80">
						<p>
							We couldn't create a passkey. If you cancelled the request to create a passkey, you
							may try again.
						</p>
					</div>
				</div>
			</div>
		</div>
	{/if}

	<form action={auth_credentials_path} method="post" onsubmit={submit}>
		<input type="hidden" name="credential" value="" />
		<input type="hidden" name="rp_name" value={data.rp.name} />
		<input type="hidden" name="state" value={data.state} />

		<button
			type="submit"
			class="inline-flex w-full cursor-pointer items-center justify-center gap-x-2 rounded-md bg-indigo-600 px-4 py-2.5 text-sm/6 font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed dark:bg-indigo-500 dark:shadow-none dark:hover:bg-indigo-400 dark:focus-visible:outline-indigo-500"
			disabled={busy}
		>
			<IconFingerprint class="size-5" />
			Create passkey
		</button>
	</form>
</div>
