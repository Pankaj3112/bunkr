// internal/recipe/compose.go
package recipe

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type composeFile struct {
	Services map[string]composeService `yaml:"services"`
	Volumes  map[string]interface{}    `yaml:"volumes,omitempty"`
}

type composeService struct {
	Image       string            `yaml:"image"`
	Ports       []string          `yaml:"ports,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Restart     string            `yaml:"restart"`
}

func GenerateCompose(r *Recipe, values map[string]string, hostPort int) ([]byte, error) {
	cf := composeFile{
		Services: make(map[string]composeService),
		Volumes:  make(map[string]interface{}),
	}

	// Primary service
	primary := composeService{
		Image:   r.Image,
		Restart: "unless-stopped",
	}

	if len(r.Ports) > 0 {
		primary.Ports = []string{fmt.Sprintf("127.0.0.1:%d:%d", hostPort, r.Ports[0])}
	}

	if len(r.Volumes) > 0 {
		primary.Volumes = r.Volumes
		for _, v := range r.Volumes {
			volName := strings.Split(v, ":")[0]
			cf.Volumes[volName] = nil
		}
	}

	if len(r.Environment) > 0 {
		env := make(map[string]string)
		for k, v := range r.Environment {
			env[k] = expandEnvValue(v, values)
		}
		primary.Environment = env
	}

	var dependsOn []string
	for _, svc := range r.Services {
		dependsOn = append(dependsOn, svc.Name)
	}
	if len(dependsOn) > 0 {
		primary.DependsOn = dependsOn
	}

	cf.Services[r.Name] = primary

	// Additional services
	for _, svc := range r.Services {
		s := composeService{
			Image:   svc.Image,
			Restart: "unless-stopped",
		}
		if len(svc.Environment) > 0 {
			s.Environment = svc.Environment
		}
		if len(svc.Volumes) > 0 {
			s.Volumes = svc.Volumes
			for _, v := range svc.Volumes {
				volName := strings.Split(v, ":")[0]
				cf.Volumes[volName] = nil
			}
		}
		cf.Services[svc.Name] = s
	}

	if len(cf.Volumes) == 0 {
		cf.Volumes = nil
	}

	data, err := yaml.Marshal(cf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate compose file: %w", err)
	}
	return data, nil
}

func expandEnvValue(value string, values map[string]string) string {
	result := value
	for k, v := range values {
		result = strings.ReplaceAll(result, "${"+k+"}", v)
	}
	return result
}
