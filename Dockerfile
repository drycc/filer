ARG LDFLAGS CODENAME

FROM registry.drycc.cc/drycc/go-dev:latest AS build
ADD . /workspace
RUN export GO111MODULE=on \
  && cd /workspace \
  && CGO_ENABLED=0 init-stack go build -ldflags "${LDFLAGS}" -o /usr/local/bin/pingguard main.go

FROM registry.drycc.cc/drycc/base:${CODENAME}

ARG DRYCC_UID=1001 \
  DRYCC_GID=1001 \
  DRYCC_HOME_DIR=/workspace \
  RCLONE_VERSION=1.72.0

RUN groupadd drycc --gid ${DRYCC_GID} \
  && useradd drycc -u ${DRYCC_UID} -g ${DRYCC_GID} -s /bin/bash -m -d ${DRYCC_HOME_DIR} \
  && install-stack rclone $RCLONE_VERSION

COPY --from=build /usr/local/bin/pingguard /usr/bin/pingguard

USER ${DRYCC_UID}
WORKDIR ${DRYCC_HOME_DIR}

ENTRYPOINT ["init-stack", "/usr/bin/pingguard"]

EXPOSE 8100
