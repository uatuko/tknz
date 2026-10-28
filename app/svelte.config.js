import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: '../.static/app',
			assets: '../.static/app',
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
