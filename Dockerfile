FROM gcr.io/distroless/base
COPY dist/assumer /assumer
ENTRYPOINT ["/assumer"]
