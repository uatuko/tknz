<script>
	/** @type {{ class: string | string[] | undefined }} */
	let { class: cls = '' } = $props();
</script>

<svg
	width="100%"
	height="100%"
	viewBox="0 0 512 512"
	version="1.1"
	xmlns="http://www.w3.org/2000/svg"
	xmlns:xlink="http://www.w3.org/1999/xlink"
	xml:space="preserve"
	style="fill-rule:evenodd;clip-rule:evenodd;stroke-linejoin:round;stroke-miterlimit:2;"
	class={cls}
>
	<rect id="shape" x="0" y="0" width="512" height="512" style="fill:none;" />
	<clipPath id="_clip1">
		<rect x="0" y="0" width="512" height="512" />
	</clipPath>
	<g clip-path="url(#_clip1)">
		<path
			d="M437,118.146c-0,-4.499 -3.647,-8.146 -8.146,-8.146l-406.803,-0c-12.17,-0 -22.051,-9.881 -22.051,-22.051l0,-65.898c0,-12.17 9.881,-22.051 22.051,-22.051l467.898,-0c12.17,-0 22.051,9.881 22.051,22.051l0,239.636c0,12.318 -10,22.318 -22.318,22.318l-467.364,0c-12.318,0 -22.318,-10 -22.318,-22.318l0,-65.364c0,-12.317 10,-22.318 22.318,-22.318l406.697,0c4.41,0 7.985,-3.575 7.985,-7.984c0,-12.245 0,-35.534 0,-47.875Zm75,274.582l0,88.554c0,16.957 -13.766,30.723 -30.723,30.723l-450.554,0c-16.957,0 -30.723,-13.766 -30.723,-30.723l0,-88.554c0,-16.956 13.766,-30.723 30.723,-30.723l450.554,0c16.957,0 30.723,13.767 30.723,30.723Z"
		/>
	</g>
</svg>
