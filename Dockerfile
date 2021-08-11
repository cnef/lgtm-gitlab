FROM alpine:3.12

COPY dist/ /

RUN chmod +x lgtm && \
    mkdir /var/lib/lgtm

VOLUME /var/lib/lgtm
EXPOSE 8989

CMD /lgtm