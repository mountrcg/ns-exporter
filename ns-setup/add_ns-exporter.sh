  #!/bin/sh

while [[ ! "$subdomain" =~ ^[a-zA-Z0-9-]+$ ]]; do
  echo
  echo "Will include ns-exporter container"
done
secret=$(cat /proc/sys/kernel/random/uuid)

cat >> docker-compose.yml <<EOF
  ns-exporter:
    image: ns-exporter:latest
    container_name: ns-exporter
    restart: unless-stopped
    environment:
    	- NS_EXPORTER_MONGO_URI=${NS_EXPORTER_MONGO_URI:-mongodb://mongo:27017}
      	- NS_EXPORTER_MONGO_DB=${NS_EXPORTER_MONGO_DB:-ns}
      	- NS_EXPORTER_INFLUX_URI=${NS_EXPORTER_INFLUX_URI:-http://influx:8086}
     	- NS_EXPORTER_INFLUX_TOKEN=yourtoken
    	- NS_EXPORTER_LIMIT=3
      	- NS_EXPORTER_SKIP=0
	  	- NS_EXPORTER_INFLUX_ORG=nighscout
		- NS_EXPORTER_INFLUX_BUCKET=ns
    depends_on:
    	- influx
	labels:
    	- 'traefik.enable=true'
