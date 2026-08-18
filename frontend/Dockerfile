FROM nginx:1.27-alpine

COPY index.html app.js config.js styles.css /usr/share/nginx/html/

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:80 || exit 1
