make-readable: make-readable-src/*
	cd make-readable-src && deno compile --output make-readable --config tsconfig.json main.ts
	mv make-readable-src/make-readable .
