# About The Project

This is the `go-megatron` service repository. Will be hosted in `amartha-ewallet-*` gcp project as `kubernetes`.
This service will focus on data filtering and transformation.

---

## Tech Stack

### Tools

- Go 1.21 or latest
- Docker

### Api Framework

- https://github.com/labstack/echo

### Driver Packages

- Redis driver https://github.com/redis/go-redis/v9
- Logger driver https://bitbucket.org/Amartha/go-x/log
- Sql driver https://github.com/jackc/pgx

### Testing Packages

- A toolkit with common assertions https://github.com/stretchr/testify/assert
- Mock https://github.com/vektra/mockery
- Redis Mock https://github.com/go-redis/redismock/v9

### Additional Packages

- Libraries for configuration parsing https://github.com/spf13/viper
- Validator https://github.com/go-playground/validator
- Swagger doc https://github.com/swaggo/swag

---

## HOW TO RUN

Clone the project inside folder `{GO_PROJECT_DIR}/src/bitbucket.org/Amartha`

```bash
git clone git@bitbucket.org:Amartha/go-payment-api.git
```

You need `config.yaml` inside `./config` for local configuration, Please ask your teammates or lead for config file.

To run this service, you need to add gcp credentials file, please ask your teammates or lead for gcp creds file, please put to `./credentials` folder because it's already registered in `.gitignore`

```bash
cp {YOUR_CRED_PATH/amartha-ewallet-*}.json ./credentials
```

This service already uses `go.mod`. `make tidy` will simply get all dependencies.

### Easy Run Service

- Run `make run-http` - Running Http Deliveries

## Run Consumer
- Run `make run-consumer consumer={consumerName}` - specify the {consumerName} based on the consumer that you want to run

### Run Unit Test

- Run `make generate` - Generate updated interfaces file
- Run `make test` - Running test without coverage.out
- Run `make test-cover` - Running test with coverage.out

---

### Contribution guidelines

- Code review - https://amartha-confluence.atlassian.net/wiki/spaces/MIS/pages/5585600618/Code+Review+Guideline+for+Reviewer
- Project structure - https://amartha-confluence.atlassian.net/wiki/spaces/AMC/pages/5686493187/Golang+Project+Structure
- APIs Guideline - https://amartha-confluence.atlassian.net/wiki/spaces/AMC/pages/5759565844/Rest+API+Style+Guide
