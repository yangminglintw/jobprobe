package config

import (
	"os"
	"regexp"
)

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// ExpandEnvVars expands environment variables in a string.
// Variables are in the format ${VAR_NAME}.
func ExpandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := envVarPattern.FindStringSubmatch(match)[1]
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		return match
	})
}

// ExpandEnvVarsInMap expands environment variables in all string values of a map.
func ExpandEnvVarsInMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = ExpandEnvVars(v)
	}
	return result
}

// ExpandEnvVarsInAuth expands environment variables in auth configuration.
func ExpandEnvVarsInAuth(auth *Auth) {
	auth.Token = ExpandEnvVars(auth.Token)
	auth.Username = ExpandEnvVars(auth.Username)
	auth.Password = ExpandEnvVars(auth.Password)
	auth.APIKey = ExpandEnvVars(auth.APIKey)
}

// ExpandEnvVarsInConfig expands all environment variables in the config.
func ExpandEnvVarsInConfig(cfg *Config) {
	for name, env := range cfg.Environments {
		env.URL = ExpandEnvVars(env.URL)
		ExpandEnvVarsInAuth(&env.Auth)
		env.Headers = ExpandEnvVarsInMap(env.Headers)
		cfg.Environments[name] = env
	}

	for i := range cfg.Jobs {
		cfg.Jobs[i].Headers = ExpandEnvVarsInMap(cfg.Jobs[i].Headers)
		cfg.Jobs[i].Options = ExpandEnvVarsInMap(cfg.Jobs[i].Options)
		expandEnvVarsInBody(cfg.Jobs[i].Body)
	}
}

// expandEnvVarsInBody recursively expands environment variables in body.
func expandEnvVarsInBody(body map[string]any) {
	for k, v := range body {
		switch val := v.(type) {
		case string:
			body[k] = ExpandEnvVars(val)
		case map[string]any:
			expandEnvVarsInBody(val)
		}
	}
}
