.EXPORT_ALL_VARIABLES:
USE_ML_MODEL=true
ML_API_ENDPOINT=https://us-south.ml.cloud.ibm.com/ml/v4/deployments/ff32d86f-7e94-4cef-92b4-9f16d9766d6b/predictions?version=2023-04-28


build:
	go build -o pulse-api .
.PHONY: local
local: build
	./pulse-api