FROM nginx:stable-alpine
MAINTAINER Peter McConnell <peter.mcconnell@rehabstudio.com>

# add scooby proxy generator
ADD ./scooby_proxy_alpine /scooby_proxy_alpine
ADD config.json /config.json
RUN chmod +x /scooby_proxy_alpine
RUN echo "*/5 * * * * /scooby_proxy_alpine" >> /etc/crontabs/root
