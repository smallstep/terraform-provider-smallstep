default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

sweep:
	TF_ACC=1 go test ./... -sweep=""
