TARGET_DIR := integrations

all: download generate

download:
	@if [ ! -d "$(TARGET_DIR)" ]; then \
		git clone git@github.com:elastic/integrations.git; \
	else \
		git -C $(TARGET_DIR) pull; \
	fi

generate:
	go run .
