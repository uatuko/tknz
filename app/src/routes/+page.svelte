<script>
	import LogoFelk from '$lib/icons/logo-felk.svelte';
	import LogoGoogle from '$lib/icons/logo-google.svelte';

	const floating = 'translate-z-30 backdrop-blur-md rounded-md shadow-lg/30';
</script>

{#snippet preview()}
	<div
		class="flex -translate-x-10 flex-col items-center gap-4 select-none perspective-normal md:-translate-x-20 lg:-translate-x-30 xl:-translate-x-20"
	>
		<div
			class="flex min-w-90 -rotate-y-35 flex-col items-center gap-5 rounded-lg border border-gray-300 bg-gray-50/40 p-8 backdrop-blur-lg lg:min-w-100"
		>
			<div class="flex w-full flex-col gap-5">
				<div class="flex w-full flex-col gap-2">
					<div class="mr-auto px-1 py-0.5 text-sm font-medium text-gray-700 text-shadow-xs">
						Username
					</div>
					<input
						disabled
						placeholder="Your username or email address"
						class="mt-1 block w-full rounded-md border border-gray-400/70 bg-white/60 px-3 py-2 shadow sm:text-sm"
					/>
				</div>

				<div class="py-3 perspective-normal">
					<div
						class={[
							'w-full rounded-md bg-indigo-600/90 px-4 py-3 text-sm font-semibold text-white',
							floating,
						]}
					>
						Continue
					</div>
				</div>
			</div>
			<div class="flex w-full items-center justify-center gap-x-3 text-xs text-gray-500">
				<hr class="grow border-t border-gray-400" />
				<span>OR</span>
				<hr class="grow border-t border-gray-400" />
			</div>
			<div class="w-full perspective-normal">
				<div
					class={[
						'flex w-full items-center justify-center gap-2 rounded-md border border-gray-300 bg-white/50 px-4 py-3 text-sm font-semibold text-gray-900',
						floating,
					]}
				>
					<LogoGoogle class="size-5" />
					Continue with Google
				</div>
			</div>
		</div>
	</div>
{/snippet}

<div class="h-screen bg-white">
	<div class="relative isolate overflow-hidden bg-linear-to-b from-indigo-100/20">
		<div
			class="mx-auto max-w-7xl pt-10 pb-24 sm:pb-32 lg:grid lg:grid-cols-2 lg:gap-x-8 lg:px-8 lg:py-40"
		>
			<div class="px-6 lg:px-0 lg:pt-4">
				<div class="mx-auto max-w-2xl">
					<div class="max-w-lg">
						<LogoFelk class="size-10 rounded-md bg-black/90 fill-white p-2.5" />

						<h1
							class="mt-10 text-5xl font-bold tracking-tight text-pretty text-gray-900 sm:text-7xl"
						>
							Felk IDaaS
						</h1>
						<p class="mt-8 text-lg text-pretty text-gray-500 sm:text-xl/8">
							The most advanced Identity and Access Management (IAM) solution at your fingertips.
						</p>
						<div class="mt-10 flex items-center gap-x-6">
							<a
								href="https://felk.ai"
								class="inline-flex gap-x-1.5 rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
							>
								Learn more
								<svg xmlns="http://www.w3.org/2000/svg" class="size-5" viewBox="0 0 512 512"
									><path
										fill="none"
										stroke="currentColor"
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="48"
										d="M268 112l144 144-144 144M392 256H100"
									/></svg
								>
							</a>
						</div>
					</div>
				</div>
			</div>
			<div class="mt-20 sm:mt-24 md:mx-auto md:max-w-2xl lg:mx-0 lg:mt-0 lg:w-screen">
				<div
					aria-hidden="true"
					class="absolute inset-y-0 right-1/2 -z-10 -mr-10 w-[200%] skew-x-[-30deg] bg-white shadow-xl ring-1 shadow-indigo-600/10 ring-indigo-50 md:-mr-20 lg:-mr-36"
				></div>
				<div class="shadow-lg md:rounded-3xl">
					<div
						class="bg-indigo-500 [clip-path:inset(0)] md:[clip-path:inset(0_round_var(--radius-3xl))]"
					>
						<div
							aria-hidden="true"
							class="absolute -inset-y-px left-1/2 -z-10 ml-10 w-[200%] skew-x-[-30deg] bg-indigo-100 opacity-20 inset-ring inset-ring-white md:ml-20 lg:ml-36"
						></div>
						<div class="relative px-6 py-16 md:pr-0 md:pl-16">
							<div class="mx-auto max-w-2xl md:mx-0 md:max-w-none">
								{@render preview()}
							</div>
							<div
								aria-hidden="true"
								class="pointer-events-none absolute inset-0 ring-1 ring-black/10 ring-inset md:rounded-3xl"
							></div>
						</div>
					</div>
				</div>
			</div>
		</div>
		<div class="absolute inset-x-0 bottom-0 -z-10 h-24 bg-linear-to-t from-white sm:h-32"></div>
	</div>
</div>
