package config

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/0xPolygon/cdk/log"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/valyala/fasttemplate"
)

const (
	startTag = "{{"
	endTag   = "}}"
)

var (
	ErrCycleVars                 = fmt.Errorf("cycle vars")
	ErrMissingVars               = fmt.Errorf("missing vars")
	ErrUnsupportedConfigFileType = fmt.Errorf("unsupported config file type")
)

type FileData struct {
	Name    string
	Content string
}

type ConfigRender struct {
	// 0: default, 1: specific
	FilesData []FileData
	// Function to resolve environment variables typically: Os.LookupEnv
	LookupEnvFunc     func(key string) (string, bool)
	EnvinormentPrefix string
}

func NewConfigRender(filesData []FileData, envinormentPrefix string) *ConfigRender {
	return &ConfigRender{
		FilesData:         filesData,
		LookupEnvFunc:     os.LookupEnv,
		EnvinormentPrefix: envinormentPrefix,
	}
}

// - Merge all files
// - Resolve all variables inside
func (c *ConfigRender) Render() (string, error) {
	mergedData, err := c.Merge()
	if err != nil {
		return "", fmt.Errorf("fail to merge files. Err: %w", err)
	}
	return c.ResolveVars(mergedData)
}

func (c *ConfigRender) Merge() (string, error) {
	k := koanf.New(".")
	for _, data := range c.FilesData {
		dataToml := c.convertVarsToStrings(data.Content)
		err := k.Load(rawbytes.Provider([]byte(dataToml)), toml.Parser())
		if err != nil {
			log.Errorf("error loading file %s. Err:%v.FileData: %v", data.Name, err, dataToml)
			return "", fmt.Errorf("fail to load  converted template %s to toml. Err: %w", data.Name, err)
		}
	}
	marshaled, err := k.Marshal(toml.Parser())
	if err != nil {
		return "", fmt.Errorf("fail to marshal to toml. Err: %w", err)
	}
	out2 := RemoveQuotesForVars(string(marshaled))
	return out2, err
}

func (c *ConfigRender) ResolveVars(fullConfigData string) (string, error) {
	// Read values, the values that are indirections get the var string "{{tag}}"
	// this step doesn't resolve any var
	tpl, valuesDefined, err := c.ReadTemplateAdnDefinedValues(fullConfigData)
	if err != nil {
		return "", err
	}
	// It fills the defined vars, if a var is not defined keep the template form:
	// A={{B}}
	renderedTemplateWithResolverVars := c.executeTemplate(tpl, valuesDefined, true)
	renderedTemplateWithResolverVars = RemoveTypeMarks(renderedTemplateWithResolverVars)
	// ? there are unresolved vars??. This means that is not a cycle, just
	// a missing value
	unresolvedVars := c.GetUnresolvedVars(tpl, valuesDefined, true)
	if len(unresolvedVars) > 0 {
		return renderedTemplateWithResolverVars, fmt.Errorf("missing vars: %v. Err: %w", unresolvedVars, ErrMissingVars)
	}
	// If there are still vars on configfile it means there are cycles:
	// Cycles are vars that depend on each other:
	// A= {{B}} and B= {{A}}
	// Also can be bigger cycles:
	// A= {{B}} and B= {{C}} and C= {{A}}
	// The way to detect that is, after resolving all vars if there are still vars to resolve,
	// then there is a cycle
	finalConfigData, err := c.ResolveCycle(renderedTemplateWithResolverVars)
	if err != nil {
		return fullConfigData, err
	}
	return finalConfigData, err
}

// ResolveCycle resolve the cycle vars:
// It iterate to configData, each step must reduce the number of 'vars'
// if not means that there are cycle vars
func (c *ConfigRender) ResolveCycle(partialResolvedConfigData string) (string, error) {
	tmpData := RemoveQuotesForVars(partialResolvedConfigData)
	pendinVars := c.GetVars(tmpData)
	if len(pendinVars) == 0 {
		// Nothing to do resolve
		return partialResolvedConfigData, nil
	}
	log.Debugf("ResolveCycle: pending vars: %v", pendinVars)
	previousData := tmpData
	for ok := true; ok; ok = len(pendinVars) > 0 {
		previousVars := pendinVars
		tpl, valuesDefined, err := c.ReadTemplateAdnDefinedValues(previousData)
		if err != nil {
			log.Errorf("resolveCycle: fails ReadTemplateAdnDefinedValues. Err: %v. Data:%s", err, previousData)
			return "", fmt.Errorf("fails to read template ResolveCycle. Err: %w", err)
		}
		renderedTemplateWithResolverVars := c.executeTemplate(tpl, valuesDefined, true)
		tmpData = RemoveQuotesForVars(renderedTemplateWithResolverVars)
		tmpData = RemoveTypeMarks(tmpData)

		pendinVars = c.GetVars(tmpData)
		if len(pendinVars) == len(previousVars) {
			return partialResolvedConfigData, fmt.Errorf("not resolved cycle vars: %v. Err: %w", pendinVars, ErrCycleVars)
		}
		previousData = tmpData
	}
	return previousData, nil
}

// The variables in data must be in format template:
// A={{B}} no A="{{B}}"
func (c *ConfigRender) ReadTemplateAdnDefinedValues(data string) (*fasttemplate.Template,
	map[string]interface{}, error) {
	tpl, err := fasttemplate.NewTemplate(data, startTag, endTag)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to load template ReadTemplateAdnDefinedValues. Err:%w", err)
	}
	out := c.convertVarsToStrings(data)
	k := koanf.New(".")
	err = k.Load(rawbytes.Provider([]byte(out)), toml.Parser())
	if err != nil {
		return nil, nil, fmt.Errorf("error ReadTemplateAdnDefinedValues parsing"+
			" data koanf.Load.Content: %s.  Err: %w", out, err)
	}
	return tpl, k.All(), nil
}

func (c *ConfigRender) convertVarsToStrings(data string) string {
	re := regexp.MustCompile(`=\s*\{\{([^}:]+)\}\}`)
	data = re.ReplaceAllString(data, `= "{{${1}:int}}"`)
	return data
}

func RemoveQuotesForVars(data string) string {
	re := regexp.MustCompile(`=\s*\"\{\{([^}:]+:int)\}\}\"`)
	data = re.ReplaceAllStringFunc(data, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) > 1 {
			parts := strings.Split(submatch[1], ":")
			return "= {{" + parts[0] + "}}"
		}
		return match
	})
	return data
}

func RemoveTypeMarks(data string) string {
	re := regexp.MustCompile(`\{\{([^}:]+:int)\}\}`)
	data = re.ReplaceAllStringFunc(data, func(match string) string {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) > 1 {
			parts := strings.Split(submatch[1], ":")
			return "{{" + parts[0] + "}}"
		}
		return match
	})
	return data
}

func (c *ConfigRender) executeTemplate(tpl *fasttemplate.Template,
	data map[string]interface{},
	useEnv bool) string {
	return tpl.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		if useEnv {
			if v, ok := c.findTagInEnvironment(tag); ok {
				tmp := fmt.Sprintf("%v", v)
				return w.Write([]byte(tmp))
			}
		}
		if v, ok := data[tag]; ok {
			tmp := fmt.Sprintf("%v", v)
			return w.Write([]byte(tmp))
		}

		v := composeVarKeyForTemplate(tag)
		return w.Write([]byte(v))
	})
}

// GetUnresolvedVars returns the vars in template that are no on data
// In this case we don't use environment variables
func (c *ConfigRender) GetUnresolvedVars(tpl *fasttemplate.Template,
	data map[string]interface{}, useEnv bool) []string {
	var unresolved []string
	tpl.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		if useEnv {
			if v, ok := c.findTagInEnvironment(tag); ok {
				return w.Write([]byte(v))
			}
		}
		if _, ok := data[tag]; !ok {
			if !contains(unresolved, tag) {
				unresolved = append(unresolved, tag)
			}
		}
		return w.Write([]byte(""))
	})
	return unresolved
}

func contains(vars []string, search string) bool {
	for _, v := range vars {
		if v == search {
			return true
		}
	}
	return false
}

// GetVars returns the vars in template
func (c *ConfigRender) GetVars(configData string) []string {
	tpl, err := fasttemplate.NewTemplate(configData, startTag, endTag)
	if err != nil {
		return []string{}
	}
	vars := unresolvedVars(tpl, map[string]interface{}{})
	return vars
}

func (c *ConfigRender) findTagInEnvironment(tag string) (string, bool) {
	envTag := c.composeVarKeyForEnvirnonment(tag)
	if v, ok := c.LookupEnvFunc(envTag); ok {
		return v, true
	}
	return "", false
}

func (c *ConfigRender) composeVarKeyForEnvirnonment(key string) string {
	return c.EnvinormentPrefix + "_" + strings.ReplaceAll(key, ".", "_")
}

func composeVarKeyForTemplate(key string) string {
	return startTag + key + endTag
}

func readFileToString(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func unresolvedVars(tpl *fasttemplate.Template, data map[string]interface{}) []string {
	var unresolved []string
	tpl.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		if _, ok := data[tag]; !ok {
			unresolved = append(unresolved, tag)
		}
		return w.Write([]byte(""))
	})
	return unresolved
}

func convertFileToToml(fileData string, fileType string) (string, error) {
	switch strings.ToLower(fileType) {
	case "json":
		k := koanf.New(".")
		err := k.Load(rawbytes.Provider([]byte(fileData)), json.Parser())
		if err != nil {
			return fileData, fmt.Errorf("error loading json file. Err: %w", err)
		}
		config := k.Raw()
		tomlData, err := toml.Parser().Marshal(config)
		if err != nil {
			return fileData, fmt.Errorf("error converting json to toml. Err: %w", err)
		}
		return string(tomlData), nil
	case "yml", "yaml", "ini":
		return fileData, fmt.Errorf("cant convert from %s to TOML. Err: %w", fileType, ErrUnsupportedConfigFileType)
	default:
		log.Warnf("filetype %s unknown, assuming is a TOML file", fileType)
		return fileData, nil
	}
}
