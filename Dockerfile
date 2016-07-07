FROM alpine
RUN apk add --update ca-certificates
ADD . /tigerbat
WORKDIR /tigerbat
ENV UPSTREAM_SERVER http://www.example.com
ENV ETCD http://etcd:2379
EXPOSE 80 8000
CMD ./tigerbat server --address=0.0.0.0:80 --mirror-url=${UPSTREAM_SERVER} --peering-address=${HOSTNAME}:8080 --etcd=${ETCD}
