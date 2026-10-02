import adapter from '@sveltejs/adapter-auto';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			out: '.build',
		}),
	},
	compilerOptions: {
		runes: true,
	},
};

export default config;
