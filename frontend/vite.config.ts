import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
    plugins: [sveltekit()],
    server:{
        host:   '0.0.0.0',
        port:   5173,
        proxy:  {
            '/api' :{
                target: 'http://localhost:8080',
                changeOrigin: true,
                ws: true,
            },
            '/config' :{
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
            '/login' :{
                target: 'http://localhost:8080',
                changeOrigin: true,
            }

        }
    }
});
