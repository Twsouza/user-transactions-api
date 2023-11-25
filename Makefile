.PHONY: startDevServer startDevEnvironment

startDevServer:
	if [ ! -f .env ]; then \
		cp .env.example .env; \
	fi
	air -c .air.toml

startDevEnvironment:
	docker-compose -f .devcontainer/docker-compose.yml up -d
