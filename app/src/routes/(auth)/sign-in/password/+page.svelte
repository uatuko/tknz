<script>
	import { auth_cb_path } from '$lib/consts.js';

	let { data } = $props();
</script>

<form action={auth_cb_path} method="POST" class="space-y-6">
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
		<label for="password" class="block text-sm/6 font-medium text-gray-900 dark:text-white">
			Password for <span class="font-semibold">{data.loginHint} </span>
			<a
				href={data.backUri}
				class="font-semibold text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300"
				>Not you?</a
			>
		</label>

		<div class="mt-2">
			<!-- svelte-ignore a11y_autofocus -->
			<input
				id="password"
				type="password"
				name="password"
				autocomplete="current-password"
				required
				class="block w-full rounded-md bg-white px-3 py-2 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6 dark:bg-white/5 dark:text-white dark:outline-white/10 dark:placeholder:text-gray-500 dark:focus:outline-indigo-500"
				autofocus
			/>
			<input type="hidden" name="login_hint" value={data.loginHint} />
			<input type="hidden" name="state" value={data.state} />
		</div>
	</div>

	<button
		type="submit"
		class="flex w-full cursor-pointer justify-center rounded-md bg-indigo-600 px-4 py-2.5 text-sm/6 font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 dark:bg-indigo-500 dark:shadow-none dark:hover:bg-indigo-400 dark:focus-visible:outline-indigo-500"
	>
		Sign in
	</button>
</form>
