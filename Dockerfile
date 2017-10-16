FROM alpine:3.6

COPY dist/ /

RUN chmod +x lgtm && \
    mkdir /var/lib/lgtm

ENV LGTM_NOTE=LGTM \
    LGTM_COUNT=2 \
    LGTM_PORT=8989 \
    LGTM_TOKEN= \
    LGTM_GITLAB_URL=http://gitlab.com \
    LGTM_DB_PATH=/var/lib/lgtm/lgtm.data \
    LGTM_LOG_LEVEL=info

VOLUME /var/lib/lgtm
EXPOSE 8989

CMD ["sh", "-c", \
    "/lgtm -token=$LGTM_TOKEN -gitlab_url=$LGTM_GITLAB_URL -lgtm_count=$LGTM_COUNT -lgtm_note=$LGTM_NOTE -log_level=$LGTM_LOG_LEVEL -db_path=$LGTM_DB_PATH -port=$LGTM_PORT" \
]