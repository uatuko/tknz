<script>
	import './layout.css';

	import { felk_uri, privacy_uri } from '$lib/consts';

	let { children } = $props();
</script>

<div class="grid h-screen py-10">
	<main class="flex min-h-full flex-col justify-center space-y-10 px-6 py-12 lg:px-8">
		{@render children?.()}
	</main>

	<footer class="w-full place-self-end">
		<div class="mx-auto flex justify-center gap-x-6">
			<p class="text-xs/6 text-gray-500 dark:text-gray-400">
				Powered by
				<a
					target="_blank"
					href={felk_uri}
					class="font-semibold text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300"
					>Felk</a
				>
			</p>
			<p class="text-xs/6 text-gray-500 dark:text-gray-400">
				<a
					target="_blank"
					href={privacy_uri}
					class="text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-300"
				>
					Privacy Policy
				</a>
			</p>
		</div>
	</footer>
</div>
