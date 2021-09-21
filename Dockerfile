# Creating a builder container to compile go files

FROM golang:alpine as builder
RUN mkdir /build
ADD /labCode /build/
WORKDIR /build
RUN go build -o main .

# Deploying go binary from builder to lighter container

FROM alpine
ARG ENTRY
ENV ENTRY=${ENTRY}
# Need to change permissions for using icmp (and probably tcp)
# RUN adduser -S -D -H -h /app appuser
# USER appuser
COPY --from=builder /build/main /app/
WORKDIR /app
CMD if [[ -n "$ENTRY" ]]; then ./main auto entry; else ./main auto; fi
