FROM registry.indexexchange.com:5000/ubuntu:20.04 AS build
RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y wget
WORKDIR /tmp
RUN wget https://dl.google.com/go/go1.19.2.linux-amd64.tar.gz && \
    tar -xf go1.19.2.linux-amd64.tar.gz && \
    mv go /usr/local
RUN mkdir -p /app/prebid-server/
WORKDIR /app/prebid-server/
ENV GOROOT=/usr/local/go
ENV PATH=$GOROOT/bin:$PATH
ENV GOPROXY="http://nexus3.indexexchange.com/repository/go-mod-group,https://proxy.golang.org,direct"
ENV GONOPROXY="*.indexexchange.com"
ENV GOPRIVATE="*.indexexchange.com"

RUN apt-get update && \
    apt-get install -y git && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN set -xe; \
	mkdir -p ~/.ssh \
	&& wget http://build-resources.indexexchange.com:8000/self-signed-certs/containerization_read_only/id_rsa -O /root/.ssh/id_rsa \
	&& wget http://build-resources.indexexchange.com:8000/self-signed-certs/containerization_read_only/id_rsa.pub -O /root/.ssh/id_rsa.pub \
	&& chmod -R 600 /root/.ssh \
	&& printf "Host *\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config \
	&& git config --global url."git@gitlab.indexexchange.com:".insteadOf "https://gitlab.indexexchange.com/"

ENV CGO_ENABLED=0
COPY ./ ./
RUN go mod tidy
RUN go mod vendor
ARG TEST="true"
RUN if [ "$TEST" != "false" ]; then ./validate.sh ; fi
RUN go build -mod=vendor -ldflags "-X github.com/prebid/prebid-server/version.Ver=`git describe --tags | sed 's/^v//'` -X github.com/prebid/prebid-server/version.Rev=`git rev-parse HEAD`" .

FROM registry.indexexchange.com:5000/ubuntu:20.04 AS release
LABEL maintainer="hans.hjort@xandr.com" 
WORKDIR /usr/local/bin/
COPY --from=build /app/prebid-server .
RUN chmod a+xr prebid-server
COPY static static/
COPY stored_requests/data stored_requests/data
RUN chmod -R a+r static/ stored_requests/data
RUN apt-get update && \
    apt-get install -y ca-certificates mtr && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN adduser prebid_user
USER prebid_user
EXPOSE 8000
EXPOSE 6060
ENTRYPOINT ["/usr/local/bin/prebid-server"]
CMD ["-v", "1", "-logtostderr"]
