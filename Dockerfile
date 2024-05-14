FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22.3 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV GOCACHE=/root/.cache/go-build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN --mount=type=cache,target="/root/.cache/go-build" go build -ldflags="-w -s" -o sharesecrets

RUN useradd -u 10000 user


FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch

WORKDIR /app

COPY --from=builder /app/sharesecrets /app/sharesecrets
COPY --from=builder /etc/passwd /etc/passwd

USER user

ENV APP_LISTEN 0.0.0.0:8000

ENTRYPOINT ["/app/sharesecrets"]
