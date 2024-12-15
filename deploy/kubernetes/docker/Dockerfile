# syntax=docker/dockerfile:experimental

ARG BASE_IMAGE=uniffle-base:latest
FROM ${BASE_IMAGE}

ARG HADOOP_VERSION
ARG HADOOP_SHORT_VERSION
ARG RSS_VERSION
ARG AUTHOR=unknown
ARG GIT_BRANCH=unknown
ARG GIT_COMMIT=unknown
ARG HADOOP_FILE=hops-$HADOOP_VERSION.tgz

LABEL author=${AUTHOR}
LABEL git-branch=${GIT_BRANCH}
LABEL git-commit=${GIT_COMMIT}

RUN useradd -ms /bin/bash rssadmin
RUN mkdir -p /data/rssadmin/
RUN chown -R rssadmin:rssadmin /data
RUN chmod -R 777 /data
USER rssadmin

COPY --chown=rssadmin:rssadmin rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz /data/rssadmin
RUN tar --no-same-owner -xvf /data/rssadmin/rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz -C /data/rssadmin
RUN mv /data/rssadmin/rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION} /data/rssadmin/rss
RUN rm /data/rssadmin/rss/conf/rss-env.sh
RUN rm -rf /data/rssadmin/rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz

COPY --chown=rssadmin:rssadmin rss-env.sh /data/rssadmin/rss/conf

RUN chmod -R 777 /data/rssadmin/rss

COPY --chmod=755 rss-env.sh /data/rssadmin/rss/bin

COPY --chmod=755 start.sh /data/rssadmin/rss/bin

COPY --chown=rssadmin:rssadmin $HADOOP_FILE /data/rssadmin
RUN mkdir -p /data/rssadmin/hops-${HADOOP_VERSION}
RUN tar --no-same-owner -zxvf /data/rssadmin/$HADOOP_FILE --strip-components=1 -C /data/rssadmin/hops-${HADOOP_VERSION}
RUN mv /data/rssadmin/hops-${HADOOP_VERSION} /data/rssadmin/hadoop
RUN rm -rf /data/rssadmin/$HADOOP_FILE
COPY --chown=rssadmin:rssadmin hadoopconfig/ /data/rssadmin/hadoop/etc/hadoop

ENV RSS_VERSION ${RSS_VERSION}

WORKDIR /data/rssadmin
CMD ["bash", "/data/rssadmin/rss/bin/start.sh"]
