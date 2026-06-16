<script>
	import LogoGoogle from '$lib/icons/logo-google.svelte';

	import { auth_sign_in_path, provider_slug_google_oauth, sign_up_path } from '$lib/consts';

	let { data } = $props();
</script>

{#await data.providers}
	<div class="flex justify-center py-8">
		<svg
			class="size-5 animate-spin text-gray-900 dark:text-gray-100"
			xmlns="http://www.w3.org/2000/svg"
			fill="none"
			viewBox="0 0 24 24"
		>
			<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
			></circle>
			<path
				class="opacity-75"
				fill="currentColor"
				d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
			></path>
		</svg>
	</div>
{:then providers}
	{#if providers.use_login}
		<form action={auth_sign_in_path} method="post" class="space-y-6">
			<div>
				<label for="login_hint" class="block text-sm/6 font-medium text-gray-900 dark:text-white"
					>Username</label
				>
				<div class="mt-2">
					<input
						id="login_hint"
						type="text"
						name="login_hint"
						required
						placeholder="Your username or email address"
						value={data.loginHint}
						class="block w-full rounded-md bg-white px-3 py-2 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6 dark:bg-white/5 dark:text-white dark:outline-white/10 dark:placeholder:text-gray-500 dark:focus:outline-indigo-500"
					/>
					<input type="hidden" name="state" value={data.state} />
				</div>
			</div>

			<button
				type="submit"
				class="inline-flex w-full cursor-pointer justify-center rounded-md bg-indigo-600 px-4 py-2.5 text-sm/6 font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 dark:bg-indigo-500 dark:shadow-none dark:hover:bg-indigo-400 dark:focus-visible:outline-indigo-500"
				>Continue</button
			>
		</form>
	{/if}
	{#if providers.oidc.length > 0}
		<div class="space-y-4">
			{#if providers.use_login}
				<div class="flex items-center gap-x-6">
					<div class="w-full flex-1 border-t border-gray-200 dark:border-white/10"></div>
					<p class="text-sm/6 text-nowrap text-gray-500 dark:text-gray-400">OR</p>
					<div class="w-full flex-1 border-t border-gray-200 dark:border-white/10"></div>
				</div>
			{/if}

			<div class="grid grid-cols-1 gap-4">
				{#each providers.oidc as p (p.id)}
					<form action={p.authorization_endpoint} method="get" class="w-full">
						<input type="hidden" name="response_type" value="code" />
						<input type="hidden" name="scope" value="openid email" />
						<input type="hidden" name="redirect_uri" value={p.redirect_uri} />

						<input type="hidden" name="client_id" value={p.client_id} />
						<input type="hidden" name="login_hint" value={data.loginHint} />
						<input type="hidden" name="state" value={data.state} />

						{#if p.slug === provider_slug_google_oauth}
							<button
								type="submit"
								class="flex w-full cursor-pointer items-center justify-center gap-3 rounded-md bg-white px-4 py-2.5 text-sm font-semibold text-gray-900 shadow-xs inset-ring inset-ring-gray-300 hover:bg-gray-50 focus-visible:inset-ring-transparent focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 dark:bg-white/10 dark:text-white dark:shadow-none dark:inset-ring-white/5 dark:hover:bg-white/20 dark:focus-visible:outline-indigo-500"
							>
								<LogoGoogle class="size-5" />
								<span class="text-sm/6 font-semibold">Continue with Google</span>
							</button>
						{/if}
					</form>
				{/each}
			</div>
		</div>
	{/if}
	{#if providers.sign_up}
		<form action={sign_up_path} method="get">
			<input type="hidden" name="state" value={data.state} />
			<input type="hidden" name="provider_id" value={providers.sign_up.id} />

			<p class="mt-10 text-center text-sm/6 text-gray-500 dark:text-gray-400">
				Don&rsquo;t have an account?
				<button
					type="submit"
					class="cursor-pointer font-semibold text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300"
					>Sign up</button
				>
			</p>
		</form>
	{/if}
{:catch}
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
				<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Request failed</h3>
				<div class="mt-2 text-sm text-red-700 dark:text-red-200/80">
					<p>
						We couldn&rsquo;t complete your sign-in request. This could be due to an invalid or
						malformed request, or due to a temporary issue.
					</p>
				</div>
			</div>
		</div>
	</div>
{/await}
