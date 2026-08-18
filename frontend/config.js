// Runtime configuration for the static frontend. Kept as a plain global
// (rather than a build-time env var) so this file can be bind-mounted or
// swapped per-environment without a build step.
window.APP_CONFIG = {
  GO_API_BASE_URL: 'http://localhost:8080',
};
