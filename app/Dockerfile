FROM golang

WORKDIR /build/
COPY ./app/src/ ./
RUN go mod download
RUN go build -ldflags '-linkmode external -extldflags "-static"' \
             -o ./main "."

FROM scratch

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /etc/passwd /etc/group /etc/
COPY --from=0 /build/main /

USER nobody:nogroup
STOPSIGNAL SIGINT

# ENV HEALTHCHECK_ENABLE 1
# HEALTHCHECK CMD healthy http://127.0.0.1:8000/ping || exit 1

ENTRYPOINT ["/main"]
EXPOSE 8000
