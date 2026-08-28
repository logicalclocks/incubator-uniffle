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
COPY rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz $HADOOP_FILE hadoop-keep-jars.txt ./

# The Hadoop tarball is a full distribution; RSS reaches a small fraction of it, so the rest
# is pruned here rather than in the final stage. Pruning in this throwaway stage means the
# removed files never enter a shipped layer at all - a `RUN rm` after the COPY below would
# leave them in the COPY layer and make the image larger, not smaller.
#
# What is removed: static archives (.a) and the YARN webapp .war, which no JVM can load;
# source and test jars; and every remaining jar not matched by hadoop-keep-jars.txt, which
# is the jdeps reachability closure for the coordinator/shuffle-server/dashboard/cli.
RUN set -eu; \
    mkdir -p rss hadoop; \
    tar --no-same-owner -xf rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz --strip-components=1 -C rss; \
    tar --no-same-owner -zxf $HADOOP_FILE --strip-components=1 -C hadoop; \
    rm rss/conf/rss-env.sh; \
    cd hadoop; \
    find . \( -name '*.a' -o -name '*.war' -o -name '*sources.jar' -o -name '*tests.jar' \) -delete; \
    rm -rf share/hadoop/*/sources share/hadoop/*/test share/hadoop/*/*/test; \
    export LC_ALL=C; \
    find share -name '*.jar' | sort > /tmp/have.txt; \
    grep -vE '^[[:space:]]*(#|$)' ../hadoop-keep-jars.txt > /tmp/patterns.txt; \
    : > /tmp/keep.txt; \
    while read -r pat; do find share -path "$pat" >> /tmp/keep.txt; done < /tmp/patterns.txt; \
    sort -u -o /tmp/keep.txt /tmp/keep.txt; \
    comm -23 /tmp/have.txt /tmp/keep.txt | xargs -r rm -f; \
    find . -type d -empty -delete || true; \
    echo "hadoop dist pruned to $(find . -name '*.jar' | wc -l) jars, $(du -sm . | cut -f1) MB"



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

# pebble is Canonical's service manager, inherited from the Ubuntu base image. Nothing in
# this image references it - the entrypoint is start.sh - and it carries its own CVEs.
RUN rm -f /usr/bin/pebble

USER rssadmin

ENV RSS_VERSION ${RSS_VERSION}

WORKDIR /data/rssadmin
CMD ["bash", "/data/rssadmin/rss/bin/start.sh"]
