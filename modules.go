package irc

import "os"

/*
Handles module loading, unloading, and registration for connections.
*/

//Module struct for defining modules
type Module struct {
	Callbacks map[string]Callback
	Info      string
	Name      string
	loaded    bool
	OnLoad    func(*Connection)
	OnUnload  func(*Connection)
}

//Loads a module and calls it's callback
func (conn *Connection) loadModule(module *Module) (err os.Error) {
	if module.loaded {
		return os.NewError("Module already loaded: " + module.Name)
	}
	for cmd, call := range module.Callbacks {
		conn.AddCallback(cmd, call)
	}
	module.loaded = true

	if module.OnLoad != nil {
		module.OnLoad(conn)
	}
	return nil
}

//Unloads a module and calls it's callback
func (conn *Connection) unloadModule(module *Module) (err os.Error) {
	if !module.loaded {
		return os.NewError("Module not loaded: " + module.Name)
	}
	for cmd, call := range module.Callbacks {
		conn.DelCallback(cmd, call)
	}
	module.loaded = false

	if module.OnUnload != nil {
		module.OnUnload(conn)
	}
	return nil
}

//Loads a module
func (conn *Connection) Load(mod string) (err os.Error) {
	module, exists := conn.modules[mod]
	if !exists {
		return os.NewError("Unknown module: " + mod)
	}
	return conn.loadModule(module)
}

//Unloads a module
func (conn *Connection) Unload(mod string) (err os.Error) {
	module, exists := conn.modules[mod]
	if !exists {
		return os.NewError("Unknown module: " + mod)
	}
	return conn.unloadModule(module)
}

//Register a module for loading
func (conn *Connection) RegisterModule(module *Module) (err os.Error) {
	_, exists := conn.modules[module.Name]
	if exists {
		return os.NewError("Module already registered by that name: " + module.Name)
	}
	conn.modules[module.Name] = module
	return nil
}


//Registers all the modules in the list passed. Modules will only be registered
//if every module in the list is registered sucessfully
func (conn *Connection) RegisterModules(modules []*Module) (err os.Error) {
	for i, module := range modules {
		if err := conn.RegisterModule(module); err != nil {
			//Unregister them because we had an error
			for _, mod := range modules[:i] {
				conn.modules[mod.Name] = nil, false
			}
			return err
		}
	}
	return nil
}
