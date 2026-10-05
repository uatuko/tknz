import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { parseArgs } from 'node:util';
import { writeFile } from 'node:fs/promises';

import pb from 'protobufjs';

/** @param {string | undefined} msg */
function usage(msg) {
	if (msg) {
		console.error('error: %s', msg);
	}

	const fname = path.basename(fileURLToPath(import.meta.url));
	console.error(`
Usage: node ${fname} [options...] <files...>
  -protodir string
        import path for proto files
  -out string
        output file for generated JSON descriptors (default "out.json")

Examples:
  node ${fname} --protodir=/path/to/proto --out=/path/to/svelte/src/lib/server/proto/example.json tknz/v1/example.proto
`);
}

async function main() {
	const { values, positionals } = parseArgs({
		options: {
			protodir: {
				type: 'string',
			},
			out: {
				type: 'string',
				default: 'out.json',
			},
		},
		allowPositionals: true,
	});

	if (!values.protodir) {
		return usage('missing proto dir');
	}

	if (positionals.length === 0) {
		return usage('no proto files');
	}

	const protodir = path.resolve(values.protodir);
	const out = path.resolve(values.out);

	const root = new pb.Root();
	root.resolvePath = function (_, target) {
		return path.join(protodir, target);
	};

	for (const file of positionals) {
		root.loadSync(file);
	}

	console.log(' ░░░░░░░░░░░░░░ | writing descriptors to "%s"', out);
	await writeFile(out, JSON.stringify(root.toJSON(), null, '  '));

	console.log('\n ✓ done');
}

await main();
