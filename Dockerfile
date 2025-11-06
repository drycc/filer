ARG LDFLAGS CODENAME

FROM registry.drycc.cc/drycc/go-dev:latest AS build
ADD . /workspace
RUN export GO111MODULE=on \
  && cd /workspace \
  && CGO_ENABLED=0 init-stack go build -ldflags "${LDFLAGS}" -o /usr/local/bin/filer cmd/filer.go

FROM registry.drycc.cc/drycc/base:${CODENAME}

ARG DRYCC_UID=1001 \
  DRYCC_GID=1001 \
  DRYCC_HOME_DIR=/workspace

RUN groupadd drycc --gid ${DRYCC_GID} \
  && useradd drycc -u ${DRYCC_UID} -g ${DRYCC_GID} -s /bin/bash -m -d ${DRYCC_HOME_DIR}

COPY --from=build /usr/local/bin/filer /usr/bin/filer

USER ${DRYCC_UID}
WORKDIR ${DRYCC_HOME_DIR}

EXPOSE 8100
