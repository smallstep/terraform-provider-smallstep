default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC_LOG=INFO TF_ACC=1 go test ./... -v -timeout 20m

sweep:
	TF_ACC_LOG=INFO TF_ACC=1 go test ./internal/provider/... -v -timeout 10m -sweep="1"

generate-docs:
	go generate

generate-client:
	go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config ./internal/apiclient/v20250101-client-gen.yml ../api/versions/20250101.yml