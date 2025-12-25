// src/types/global.d.ts
interface Window {
	// 声明 msgpack5 类型（函数，返回包含 encode 方法的对象）
	msgpack5?: () => {
		encode: (data: unknown) => Uint8Array;
		decode: (buffer: Uint8Array) => unknown;
	};
}

// 声明 Buffer 全局类型（如果需要）
declare global {
	interface Window {
		Buffer?: typeof Buffer;
	}
}
