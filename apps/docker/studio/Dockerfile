FROM nginx:stable-alpine

RUN sed -i '1idaemon off;' /etc/nginx/nginx.conf

COPY docker/studio/studio.conf /etc/nginx/conf.d/default.conf

# adapt the `dist/` folder to the output directory your build tool uses (such as `dist/`, `build/` or `www/`).
COPY packages/app/dist/spa/ /app

EXPOSE 80

CMD ["nginx"]