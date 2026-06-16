import globals from 'globals';
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';

import { includeIgnoreFile } from '@eslint/compat';
import { fileURLToPath } from 'node:url';

import svelteConfig from './app/svelte.config.js';

const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url));

/** @type {import('eslint').Linter.Config[]} */
export default [
	includeIgnoreFile(gitignorePath),
	js.configs.recommended,
	...svelte.configs.recommended,
	{
		languageOptions: {
			globals: { ...globals.browser, ...globals.node },
		},
	},
	{
		files: ['**/*.svelte', '**/*.svelte.js'],
		languageOptions: { parserOptions: { svelteConfig } },
	},
	{
		rules: {
			'arrow-parens': ['error', 'always'],
			'brace-style': ['error', '1tbs', { allowSingleLine: true }],
			'comma-dangle': ['error', 'always-multiline'],
			'eol-last': ['error', 'always'],
			eqeqeq: ['error', 'always'],
			'function-call-argument-newline': ['error', 'consistent'],
			indent: ['error', 'tab', { SwitchCase: 1 }],
			'key-spacing': [
				'error',
				{
					beforeColon: false,
					afterColon: true,
					mode: 'minimum',
				},
			],
			'keyword-spacing': ['error', { before: true, after: true }],
			'linebreak-style': ['error', 'unix'],
			'no-multi-spaces': ['error'],
			'no-multiple-empty-lines': ['error', { max: 3, maxEOF: 0 }],
			'no-trailing-spaces': ['error'],
			'no-var': ['error'],
			'object-curly-newline': ['error'],
			'object-curly-spacing': ['error', 'always'],
			'object-shorthand': ['error', 'always'],
			quotes: ['error', 'single'],
			'semi-spacing': ['error', { before: false, after: false }],
			'space-in-parens': ['error'],
			'space-unary-ops': [
				'error',
				{
					words: true,
					nonwords: false,
					overrides: { typeof: false },
				},
			],
			'svelte/no-navigation-without-resolve': [
				'error',
				{
					ignoreGoto: true,
					ignoreLinks: true,
					ignorePushState: false,
					ignoreReplaceState: false,
				},
			],
		},
	},
];
