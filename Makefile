default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC_LOG=INFO TF_ACC=1 go test ./... -v -timeout 20m

sweep:
	TF_ACC_LOG=INFO TF_ACC=1 go test ./... -v -timeout 10m -sweep="1"

generate-docs:
	go generate
