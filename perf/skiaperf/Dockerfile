FROM gcr.io/skia-public/base-cipd:testing-slim

COPY . /

USER skia

ENTRYPOINT ["/usr/local/bin/skiaperf"]
CMD ["--logtostderr"]
