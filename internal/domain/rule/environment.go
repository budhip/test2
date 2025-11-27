package rule

import "fmt"

type Environment struct {
	value string
}

const (
	EnvDev  = "dev"
	EnvUat  = "uat"
	EnvProd = "prod"
)

func NewEnvironment(env string) (Environment, error) {
	validEnvs := map[string]bool{
		EnvDev:  true,
		EnvUat:  true,
		EnvProd: true,
	}

	if !validEnvs[env] {
		return Environment{}, fmt.Errorf("invalid environment: must be dev, uat, or prod")
	}

	return Environment{value: env}, nil
}

func (e Environment) String() string {
	return e.value
}

func (e Environment) Equals(other Environment) bool {
	return e.value == other.value
}

func (e Environment) IsProd() bool {
	return e.value == EnvProd
}

func (e Environment) IsDev() bool {
	return e.value == EnvDev
}
