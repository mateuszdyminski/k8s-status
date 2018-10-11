FROM alpine:3.8

RUN mkdir -p /usr/share/k8status

WORKDIR /usr/share/k8status

ADD build/k8status .
ADD config .

ENV K8STATUS_HTTPPORT=8080 
ENV K8STATUS_GRACEFULSHUTDOWNTIMEOUT=5 
ENV K8STATUS_GRACEFULSHUTDOWNEXTRASLEEP=0
ENV K8STATUS_DEBUG=true
ENV K8STATUS_KUBECONFIGPATH="/usr/share/k8status/config"

EXPOSE 8080

ENTRYPOINT /usr/share/k8status/k8status
