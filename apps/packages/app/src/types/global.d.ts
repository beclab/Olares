// src/types/global.d.ts
interface Window {
	// Declare msgpack5 type (function returning object with encode/decode).
	msgpack5?: () => {
		encode: (data: unknown) => Uint8Array;
		decode: (buffer: Uint8Array) => unknown;
	};
}

// Declare global Buffer type (if needed).
declare global {
	interface Window {
		Buffer?: typeof Buffer;
	}
}
