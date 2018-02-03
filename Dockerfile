FROM gcr.io/distroless/base
COPY /assumer /assumer
ENTRYPOINT ["/assumer"]
