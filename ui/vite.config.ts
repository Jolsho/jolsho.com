import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vite.dev/config/
export default defineConfig({
    build: {
        outDir: '../server/dst',
        emptyOutDir: true, // forces Vite to empty it even if outside root
    },
    plugins: [react()],
});
