<script>
	import LogoFelk from '$lib/icons/logo-felk.svelte';

	let { children } = $props();
</script>

<div class="sm:mx-auto sm:w-full sm:max-w-sm">
	<LogoFelk
		class="mx-auto size-12 rounded-md bg-black fill-white p-3 dark:bg-white dark:fill-black"
	/>
	<h1 class="mt-6 text-center text-xl font-semibold text-gray-900 dark:text-white">Sign in</h1>
</div>

<div class="mx-auto w-full max-w-sm space-y-5 sm:max-w-md">
	{@render children?.()}
</div>
