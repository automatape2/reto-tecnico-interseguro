// Runtime configuration for the all-in-one Vercel deployment. The frontend
// and both APIs (Go and Node serverless functions under /api) are served
// from the same Vercel domain, so requests are same-origin - no base URL
// or CORS handling needed here.
window.APP_CONFIG = {
  GO_API_BASE_URL: '',
};
