// Runtime configuration for the static frontend. Kept as a plain global
// (rather than a build-time env var) so this file can be bind-mounted or
// swapped per-environment without a build step.
//
// Deployed on Vercel: this points at go-api running on Render. If you name
// the Render service something other than "interseguro-go-api", update the
// URL below to match before deploying.
window.APP_CONFIG = {
  GO_API_BASE_URL: 'https://interseguro-go-api.onrender.com',
};
