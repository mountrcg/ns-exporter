#!/bin/sh

while [[ ! "$subdomain" =~ ^[a-zA-Z0-9-]+$ ]]; do
  echo
  echo "Enter the name of the subdomain where your Grafana Dashboard will be available:"
  read subdomainGDB
  echo
  echo "Enter the name of the subdomain where your Influx DB will be available:"
  read subdomainIDB
done
secret=$(cat /proc/sys/kernel/random/uuid)

cat >> docker-compose.yml <<EOF

	grafana:
		image: grafana/grafana-oss:latest
		container_name: grafana
		restart: always
		depends_on:
			- influx
		volumes:
			- data-grafana:/var/lib/grafana
		labels:
		- 'traefik.enable=true'
		- 'traefik.http.routers.grafana.rule=Host(\`${subdomainGDB}.\${NS_DOMAIN}\`)'
		- 'traefik.http.routers.grafana.entrypoints=web'
		- 'traefik.http.routers.grafana.entrypoints=websecure'
		- 'traefik.http.routers.grafana.tls.certresolver=le'
		ports:
			- 3000:3000

	influx:
		image: influxdb:latest
		container_name: influx
		restart: always
		volumes:
			- data-influx:/var/lib/influxdb2
		labels:
			- 'traefik.enable=true'
			- 'traefik.http.routers.influx.rule=Host(\`${subdomainIDB}.\${NS_DOMAIN}\`)'
			- 'traefik.http.routers.influx.entrypoints=web'
			- 'traefik.http.routers.influx.entrypoints=websecure'
			- 'traefik.http.routers.influx.tls.certresolver=le'
			ports:
				- 8086:8086

volumes:
  data-grafana-ns:
  data-influx-ns:
EOF

sudo docker compose up -d

echo "After editing settings, re-launch your Nightscout by typing 'sudo docker compose up -d'"