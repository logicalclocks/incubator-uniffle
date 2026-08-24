# syntax=docker/dockerfile:experimental

ARG BASE_IMAGE=uniffle-base:latest

# Unpacked in a throwaway stage. Extracting in the final image costs 2 GB: each
# tarball stays in its own COPY layer once a later RUN deletes it, and the
# rename and the recursive chmod each copy the whole tree into another layer.
FROM ${BASE_IMAGE} AS dists

ARG HADOOP_VERSION
ARG HADOOP_SHORT_VERSION
ARG RSS_VERSION
ARG HADOOP_FILE=hops-$HADOOP_VERSION.tgz

WORKDIR /dists
COPY rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz $HADOOP_FILE ./
RUN mkdir -p rss hadoop && \
    tar --no-same-owner -xf rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz --strip-components=1 -C rss && \
    tar --no-same-owner -zxf $HADOOP_FILE --strip-components=1 -C hadoop && \
    rm rss/conf/rss-env.sh


FROM ${BASE_IMAGE}

ARG RSS_VERSION
ARG AUTHOR=unknown
ARG GIT_BRANCH=unknown
ARG GIT_COMMIT=unknown

LABEL author=${AUTHOR}
LABEL git-branch=${GIT_BRANCH}
LABEL git-commit=${GIT_COMMIT}

RUN useradd -ms /bin/bash rssadmin && \
    mkdir -p /data/rssadmin/rss && \
    chown -R rssadmin:rssadmin /data && \
    chmod -R 777 /data

COPY --from=dists --chown=rssadmin:rssadmin --chmod=777 /dists/rss /data/rssadmin/rss
COPY --from=dists --chown=rssadmin:rssadmin /dists/hadoop /data/rssadmin/hadoop

COPY --chown=rssadmin:rssadmin --chmod=777 rss-env.sh /data/rssadmin/rss/conf

COPY --chmod=755 rss-env.sh /data/rssadmin/rss/bin

COPY --chmod=755 start.sh /data/rssadmin/rss/bin

COPY --chown=rssadmin:rssadmin hadoopconfig/ /data/rssadmin/hadoop/etc/hadoop

USER rssadmin

ENV RSS_VERSION ${RSS_VERSION}

WORKDIR /data/rssadmin
CMD ["bash", "/data/rssadmin/rss/bin/start.sh"]
