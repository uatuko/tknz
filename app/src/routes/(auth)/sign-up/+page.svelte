<script>
	import LogoFelk from '$lib/icons/logo-felk.svelte';

	import { auth_sign_up_path, sign_in_path } from '$lib/consts';

	let { data } = $props();
</script>

<div class="sm:mx-auto sm:w-full sm:max-w-sm">
	<LogoFelk
		class="mx-auto size-12 rounded-md bg-black fill-white p-3 dark:bg-white dark:fill-black"
	/>
	<h1 class="mt-6 text-center text-xl font-semibold text-gray-900 dark:text-white">Sign up</h1>
</div>

<div class="mx-auto w-full max-w-sm space-y-5 sm:max-w-md">
	<form action={auth_sign_up_path} method="post" class="space-y-6">
		{#if data.error}
			<div class="flex items-center gap-4 rounded-md bg-red-50 p-4 dark:bg-red-500/15">
				<svg xmlns="http://www.w3.org/2000/svg" class="size-5 text-red-400" viewBox="0 0 512 512"
					><path
						fill="currentColor"
						d="M256 48a208 208 0 1 1 0 416a208 208 0 1 1 0-416m0 464a256 256 0 1 0 0-512a256 256 0 1 0 0 512m-81-337c-9.4 9.4-9.4 24.6 0 33.9l47 47l-47 47c-9.4 9.4-9.4 24.6 0 33.9s24.6 9.4 33.9 0l47-47l47 47c9.4 9.4 24.6 9.4 33.9 0s9.4-24.6 0-33.9l-47-47l47-47c9.4-9.4 9.4-24.6 0-33.9s-24.6-9.4-33.9 0l-47 47l-47-47c-9.4-9.4-24.6-9.4-33.9 0"
					/></svg
				>
				<p class="text-sm font-medium text-red-800 dark:text-red-200">
					{data.error}
				</p>
			</div>
		{/if}

		<div>
			<label for="login_hint" class="block text-sm/6 font-medium text-gray-900 dark:text-white"
				>Email</label
			>
			<div class="mt-2">
				<input
					id="login_hint"
					type="email"
					name="login_hint"
					required
					placeholder="Your email address"
					class="block w-full rounded-md bg-white px-3 py-2 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6 dark:bg-white/5 dark:text-white dark:outline-white/10 dark:placeholder:text-gray-500 dark:focus:outline-indigo-500"
				/>
			</div>
		</div>

		<input type="hidden" name="provider_id" value={data.providerId} />
		<input type="hidden" name="state" value={data.state} />
		<button
			type="submit"
			class="inline-flex w-full cursor-pointer justify-center rounded-md bg-indigo-600 px-4 py-2.5 text-sm/6 font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 dark:bg-indigo-500 dark:shadow-none dark:hover:bg-indigo-400 dark:focus-visible:outline-indigo-500"
			>Continue</button
		>
	</form>

	<form action={sign_in_path} method="get">
		<input type="hidden" name="state" value={data.state} />

		<p class="mt-10 text-center text-sm/6 text-gray-500 dark:text-gray-400">
			Already have an account?
			<button
				type="submit"
				class="cursor-pointer font-semibold text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300"
				>Sign in</button
			>
		</p>
	</form>
</div>
