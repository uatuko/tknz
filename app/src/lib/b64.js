/**
 *
 * @param {string} str
 * @returns {Uint8Array<ArrayBuffer>}
 */
export function decode(str) {
	// @ts-ignore (ref: https://github.com/microsoft/TypeScript/pull/61696)
	return Uint8Array.fromBase64(str, { alphabet: 'base64url' });
}
