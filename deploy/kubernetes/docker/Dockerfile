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
USER rssadmin

COPY rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz /data/rssadmin
RUN tar -xvf /data/rssadmin/rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz -C /data/rssadmin
RUN mv /data/rssadmin/rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION} /data/rssadmin/rss
RUN rm /data/rssadmin/rss/bin/rss-env.sh
RUN rm -rf /data/rssadmin/rss-${RSS_VERSION}-hadoop${HADOOP_SHORT_VERSION}.tgz

COPY rss-env.sh /data/rssadmin/rss/bin

COPY start.sh /data/rssadmin/rss/bin

COPY $HADOOP_FILE /data/rssadmin
RUN mkdir -p /data/rssadmin/hops-${HADOOP_VERSION}
RUN tar -zxvf /data/rssadmin/$HADOOP_FILE --strip-components=1 -C /data/rssadmin/hops-${HADOOP_VERSION}
RUN mv /data/rssadmin/hops-${HADOOP_VERSION} /data/rssadmin/hadoop
RUN rm -rf /data/rssadmin/$HADOOP_FILE
COPY hadoopconfig/ /data/rssadmin/hadoop/etc/hadoop

ENV RSS_VERSION ${RSS_VERSION}
ENV HADOOP_VERSION ${HADOOP_VERSION}

WORKDIR /data/rssadmin
CMD ["bash", "/data/rssadmin/rss/bin/start.sh"]
