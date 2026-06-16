import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: '../.dist/app',
			assets: '../.dist/app',
			fallback: undefined,
			precompress: false,
			strict: true,
		}),
	},
	compilerOptions: {
		runes: true,
	},
};

export default config;
