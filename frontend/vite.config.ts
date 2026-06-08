import { defineConfig } from 'vite';
import { fileURLToPath } from 'url';
import tailwindcss from '@tailwindcss/vite';
import viteReact from '@vitejs/plugin-react';
import { tanstackStart } from '@tanstack/react-start/plugin/vite';

export default defineConfig({
    server: {
        port: 3000,
    },
    resolve: {
        tsconfigPaths: true,
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    plugins: [
        viteReact(),
        tailwindcss(),
        tanstackStart(),
    ],
})