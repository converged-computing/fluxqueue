ARG GO_BASE_IMAGE=golang:1.23
ARG DISTROLESS_BASE_IMAGE=gcr.io/distroless/static:nonroot
FROM $GO_BASE_IMAGE AS builder
ARG TARGETARCH=amd64

WORKDIR /workspace
COPY ./go.mod ./go.mod
COPY ./Makefile ./Makefile
COPY ./cmd/scheduler ./cmd/scheduler
COPY ./pkg/scheduler ./pkg/scheduler

ARG TARGETARCH
RUN go mod tidy && \
    go mod vendor && \
    make scheduler GO_BUILD_ENV='CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH}'

FROM $DISTROLESS_BASE_IMAGE
WORKDIR /bin
COPY --from=builder /workspace/bin/fluxqueue-scheduler .
USER 65532:65532
ENTRYPOINT ["/bin/fluxqueue-scheduler"]