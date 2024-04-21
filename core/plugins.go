package core

import (
	"C"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ebitengine/purego"
	"github.com/vanilla-os/vib/api"
)

var openedPlugins map[string]Plugin

func LoadPlugin(name string, module interface{}, recipe *api.Recipe) (string, error) {
	if openedPlugins == nil {
		openedPlugins = make(map[string]Plugin)
	}
	pluginOpened := false
	var buildModule Plugin
	buildModule, pluginOpened = openedPlugins[name]
	if !pluginOpened {
		fmt.Println("Loading new plugin")
		buildModule = Plugin{Name: name}

		loadedPlugin, err := purego.Dlopen(fmt.Sprintf("%s/%s.so", recipe.PluginPath, name), purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
		var buildFunction func(*C.char, *C.char) string
		purego.RegisterLibFunc(&buildFunction, loadedPlugin, "BuildModule")
		buildModule.BuildFunc = buildFunction
		buildModule.LoadedPlugin = loadedPlugin
		openedPlugins[name] = buildModule
	}
	fmt.Printf("Using plugin: %s\n", buildModule.Name)
	moduleJson, err := json.Marshal(module)
	if err != nil {
		return "", err
	}
	recipeJson, err := json.Marshal(recipe)
	if err != nil {
		return "", err
	}

	res := buildModule.BuildFunc(C.CString(string(moduleJson)), C.CString(string(recipeJson)))
	if strings.HasPrefix(res, "ERROR:") {
		return "", fmt.Errorf("%s", strings.Replace(res, "ERROR: ", "", 1))
	} else {
		return res, nil
	}
}
