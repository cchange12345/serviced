// Copyright 2016 The Serviced Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package facade

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sync"
	"strings"

	"github.com/zenoss/glog"
	"github.com/control-center/serviced/datastore"
	"github.com/control-center/serviced/domain/servicedefinition"
	"github.com/control-center/serviced/domain/servicetemplate"
	"github.com/control-center/serviced/isvcs"
	"github.com/control-center/serviced/utils"
	"bytes"
)


var (
	logstashConfigLock = &sync.Mutex{}
	ErrLogstashUnchanged = errors.New("logstash config unchanged")
)

// ReloadLogstashConfig will create a new logstash configuration based on the current
// templates. If that configuration is different from the one currently used by logstash,
// then it will restart the logstash container so it can use the new filter set.
//
// This method should be called anytime the available service definitions are modified
// This method depends on the elasticsearch container being up and running.
func (f *Facade) ReloadLogstashConfig(ctx datastore.Context) error {
	// serialize updates so that we don't have different threads overwriting the same file.
	logstashConfigLock.Lock()
	defer logstashConfigLock.Unlock()

	templates, err := f.GetServiceTemplates(ctx)
	if err != nil {
		glog.Errorf("Could not write logstash configuration: %s", err)
		return err
	}
	err = writeLogstashConfiguration(templates)
	if err == ErrLogstashUnchanged {
		return nil
	} else if err != nil {
		glog.Errorf("Could not write logstash configuration: %s", err)
		return err
	} 

	glog.V(2).Infof("Starting logstash container")
	if err := isvcs.Mgr.Notify("restart logstash"); err != nil {
		glog.Errorf("Could not start logstash container: %s", err)
		return err
	}
	return nil
}

type reloadLogstashContainer func(ctx datastore.Context, f FacadeInterface) error

var LogstashContainerReloader reloadLogstashContainer = reloadLogstashContainerImpl

func reloadLogstashContainerImpl(ctx datastore.Context, f FacadeInterface) error {
	return f.ReloadLogstashConfig(ctx)
}

func resourcesDir() string {
	homeDir := os.Getenv("SERVICED_HOME")
	if len(homeDir) == 0 {
		_, filename, _, _ := runtime.Caller(1)
		return path.Clean(path.Join(path.Dir(filename), "..", "isvcs", "resources"))

	}
	return path.Clean(path.Join(homeDir, "isvcs/resources"))
}

// writeLogstashConfiguration takes a map of ServiceTemplates and writes them to the
// appropriate place in the logstash.conf.
// This is required before logstash startup
//
// This method returns nil of logstash configuration was replaced,
// ErrLogstashUnchanged if the configuration is unchanged, or other errors if there was an I/O problem
// FIXME: eventually this file should live in the DFS or the config should
// live in zookeeper to allow the agents to get to this
func writeLogstashConfiguration(templates map[string]servicetemplate.ServiceTemplate) error {
	// the definitions are a map of filter name to content
	// they are found by recursively going through all the service definitions
	filterDefs := make(map[string]string)
	for _, template := range templates {
		subFilterDefs := getFilterDefinitions(template.Services)
		for name, value := range subFilterDefs {
			filterDefs[name] = value
		}
	}

	// filters will be a syntactically correct logstash filters section
	filters := ""

	for _, template := range templates {
		filters += getFilters(template.Services, filterDefs, []string{})
	}
	newConfigFile := resourcesDir() + "/logstash/logstash.conf.new"
	err := writeLogStashConfigFile(filters, newConfigFile)
	if err != nil {
		glog.Errorf("Unable to create new logstash config file %s: %s", newConfigFile, err)
		return err
	}

	originalFile := resourcesDir() + "/logstash/logstash.conf"
	originalContents, err := ioutil.ReadFile(originalFile)
	if err != nil {
		glog.Errorf("Unable to read current logstash config file %s: %s", originalFile, err)
		return err
	}

	newContents, err := ioutil.ReadFile(newConfigFile)
	if err != nil {
		glog.Errorf("Unable to read new logstash config file %s: %s", newConfigFile, err)
		return err
	}

	// Now compare the new config to the current config, and
	// only replace the current config if they are different
	if bytes.Equal(originalContents, newContents) {
		return ErrLogstashUnchanged
	} else if err := os.Rename(newConfigFile, originalFile); err != nil {
		glog.Errorf("Unable to replace logstash config file: %s", err)
		return err
	}
	glog.Infof("Updated logstash configuration")
	return nil
}

func getFilterDefinitions(services []servicedefinition.ServiceDefinition) map[string]string {
	filterDefs := make(map[string]string)
	for _, service := range services {
		for name, value := range service.LogFilters {
			filterDefs[name] = value
		}

		if len(service.Services) > 0 {
			subFilterDefs := getFilterDefinitions(service.Services)
			for name, value := range subFilterDefs {
				filterDefs[name] = value
			}
		}
	}
	return filterDefs
}

func getFilters(services []servicedefinition.ServiceDefinition, filterDefs map[string]string, typeFilter []string) string {
	filters := ""
	for _, service := range services {
		for _, config := range service.LogConfigs {
			for _, filtName := range config.Filters {
				//do not write duplicate types, logstash doesn't handle this
				if !utils.StringInSlice(config.Type, typeFilter) {
					filters += fmt.Sprintf("\nif [type] == \"%s\" \n {\n  %s \n}", config.Type, filterDefs[filtName])
					typeFilter = append(typeFilter, config.Type)
				}
			}
		}
		if len(service.Services) > 0 {
			subFilts := getFilters(service.Services, filterDefs, typeFilter)
			filters += subFilts
		}
	}
	return filters
}

// This method writes out the config file for logstash. It uses
// the logstash.conf.template and does a variable replacement.
func writeLogStashConfigFile(filters string, outputPath string) error {
	// read the log configuration template
	templatePath := resourcesDir() + "/logstash/logstash.conf.template"

	contents, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return err
	}

	// the newlines in the string below are deliberate so that the filter
	// syntax is correct
	filterSection := `
filter {
  mutate {
    rename => {
      "source" => "file"
    }
  }
# NOTE the filters are generated from the service definitions
` + string(filters) + `
}
`
	newContents := strings.Replace(string(contents), "${FILTER_SECTION}", filterSection, 1)
	newBytes := []byte(newContents)
	// generate the filters section
	// write the log file
	return ioutil.WriteFile(outputPath, newBytes, 0644)
}