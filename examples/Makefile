GO := $(shell which go)

example-gcp:
ifndef PROJECT_ID
	$(error missing GCP_PROJECT_ID)
endif
ifndef SECRET_ID
	$(error missing SECRET_ID)
endif
	$(GO) run gcp/*.go \
		./config.json \
		projects/$(PROJECT_ID)/secrets/$(SECRET_ID)
