// @ts-check
import fs from 'node:fs/promises';

import { glob } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'astro/config';

import tailwindcss from '@tailwindcss/vite';

/** @returns {import('astro').AstroIntegration} */
function htmlmin() {
	return {
		name: 'htmlmin',
		hooks: {
			'astro:build:done': async ({ dir }) => {
				for await (const fname of glob(`${fileURLToPath(dir)}**/*.html`)) {
					// WARN: this will break any preformatted text

					let out = '';
					const lines = (await fs.readFile(fname)).toString().split('\n');
					for (const line of lines) {
						out += line.trim();
						if (!out.endsWith('>')) {
							out += ' ';
						}
					}

					await fs.writeFile(fname, out);
				}
			},
		},
	};
}

// https://astro.build/config
export default defineConfig({
	build: {
		format: 'preserve',
	},
	integrations: [htmlmin()],
	outDir: '../.dist/tmpl',
	trailingSlash: 'always',
	vite: {
		build: {
			rollupOptions: {
				output: {
					assetFileNames: () => {
						return '_a/[hash].[ext]';
					},
				},
			},
			target: 'chrome58',
		},
		plugins: [tailwindcss()],
	},
});
