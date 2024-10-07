# SPDX-FileCopyrightText: Josef Andersson
#
# SPDX-License-Identifier: CC0-1.0

FROM cgr.dev/chainguard/glibc-dynamic:latest-dev
ARG TARGETOS TARGETARCH
ARG DIRPATH=""

COPY ${DIRPATH}gommitlint-${TARGETOS}-${TARGETARCH} /usr/bin/gommitlint
ENTRYPOINT ["/usr/bin/gommitlint"]
