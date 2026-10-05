<script>
	import { resolve } from '$app/paths';
	import { sign_out_path } from '$lib/consts';
</script>

<div class="flex min-h-screen items-center justify-center bg-white dark:bg-gray-900">
	<main class="w-full max-w-7xl px-6 pt-10 pb-16 sm:pb-24 lg:px-8">
		<div class="mx-auto mt-20 max-w-2xl text-center sm:mt-24">
			<h1
				class="mt-4 text-5xl font-semibold tracking-tight text-balance text-gray-900 sm:text-6xl dark:text-white"
			>
				Welcome!
			</h1>
			<p class="mt-6 text-lg text-pretty text-gray-500 sm:text-xl/8 dark:text-gray-400">
				You are signed-in.
			</p>
		</div>
		<div class="mx-auto mt-16 flow-root max-w-lg sm:mt-20">
			<div class="mt-10 flex justify-center">
				<a
					data-sveltekit-preload-data="tap"
					href={resolve(sign_out_path)}
					class="text-sm/6 font-semibold text-indigo-600 dark:text-indigo-400"
					>Sign out <span aria-hidden="true">&rarr;</span></a
				>
			</div>
		</div>
	</main>
</div>
