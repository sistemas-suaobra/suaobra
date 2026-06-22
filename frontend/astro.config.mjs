import { defineConfig } from 'astro/config';
import react from "@astrojs/react";
import sentry from "@sentry/astro";

const sentryDsn =
  "https://23f9fe3008439fdbc379a96e5af1d8ba@o4506325509931008.ingest.sentry.io/4506325524807680";

const sentryAuthToken = process.env.SENTRY_AUTH_TOKEN;
const disableSentryUpload = process.env.DISABLE_SENTRY_UPLOAD === "1";

const integrations = [react()];

if (!disableSentryUpload && sentryAuthToken) {
  integrations.push(
    sentry({
      dsn: sentryDsn,
      sourceMapsUploadOptions: {
        project: "sua-obra",
        authToken: sentryAuthToken,
      },
    }),
  );
} else {
  integrations.push(
    sentry({
      dsn: sentryDsn,
    }),
  );
}

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
  integrations,
});