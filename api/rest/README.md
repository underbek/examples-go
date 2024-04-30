# Rest

```shell
PACKAGE_VERSION=0.0.1
PACKAGE_NAME=app

java -jar ./openapi-generator-cli-7.0.1.jar generate -g python -i ./openapi/test/configuration-service/openapi.yaml -o ./gen --package-name ${PACKAGE_NAME}_api --additional-properties=npmName=${PACKAGE_NAME}_api --additional-properties=npmVersion=${PACKAGE_VERSION}

# -g typescript-axios
```
