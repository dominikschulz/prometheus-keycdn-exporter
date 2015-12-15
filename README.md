# KeyCDN exporter

The KeyCDN exporter allows exporting stats from the KeyCDN
to be exported to Prometheus.

## Building and running

### Local Build

    make
    ./keycdn_exporter <flags>

### Building with Docker

    docker build -t keycdn_exporter .
    docker run -d -p 9116:9116 --name keycdn_exporter -v `pwd`:/config keycdn_exporter -config.file=/config/keycdn.yml

## Configuration

A configuration showing all options is below:
```
apikey: "your key"
```

