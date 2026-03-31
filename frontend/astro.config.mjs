import { defineConfig } from 'astro/config';
import react from "@astrojs/react";
import sentry from "@sentry/astro";

export default defineConfig({
  vite: {
    server: {
      host: true,
      allowedHosts: [
        "app.suaobra.test",
        "api.suaobra.test",
      ],
    },
    ssr: {
      noExternal: ['primereact'],
    },
  },
  integrations: [
    react(),
    sentry({
      dsn: "https://23f9fe3008439fdbc379a96e5af1d8ba@o4506325509931008.ingest.sentry.io/4506325524807680",
      sourceMapsUploadOptions: {
        project: "sua-obra",
        authToken: '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',
      },
    }),
  ],
});