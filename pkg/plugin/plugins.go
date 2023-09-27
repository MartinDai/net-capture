package plugin

import (
	"net-capture/pkg/input"
	"net-capture/pkg/message"
	"net-capture/pkg/model"
	"net-capture/pkg/output"
	"reflect"
)

// InOutPlugins struct for holding references to plugins
type InOutPlugins struct {
	Inputs  []message.PluginReader
	Outputs []message.PluginWriter
	All     []interface{}
}

// Plugins holds all the plugin objects
//var Plugins = new(InOutPlugins)

// Automatically detects type of plugin and initialize it
//
// See this article if curious about reflect stuff below: http://blog.burntsushi.net/type-parametric-functions-golang
func (plugins *InOutPlugins) registerPlugin(constructor interface{}, options ...interface{}) {
	vc := reflect.ValueOf(constructor)

	// Pre-processing options to make it work with reflect
	var vo []reflect.Value
	for _, oi := range options {
		vo = append(vo, reflect.ValueOf(oi))
	}

	// Calling our constructor with list of given options
	plugin := vc.Call(vo)[0].Interface()

	if r, ok := plugin.(message.PluginReader); ok {
		plugins.Inputs = append(plugins.Inputs, r)
	}
	if w, ok := plugin.(message.PluginWriter); ok {
		plugins.Outputs = append(plugins.Outputs, w)
	}
	plugins.All = append(plugins.All, plugin)
}

// InitPlugins specify and initialize all available plugins
func InitPlugins(inputConfig []model.InputConfig) *InOutPlugins {
	plugins := new(InOutPlugins)

	for _, i := range inputConfig {
		plugins.registerPlugin(input.NewIPInput, i.Address)
	}

	plugins.registerPlugin(output.NewStdOutput)

	return plugins
}
