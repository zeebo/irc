package irc

import "errors"

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
func (conn *Connection) loadModule(module *Module) (err error) {
	if module.loaded {
		return errors.New("Module already loaded: " + module.Name)
	}

	module.loaded = true

	if module.OnLoad != nil {
		module.OnLoad(conn)
	}
	return nil
}

//Unloads a module and calls it's callback
func (conn *Connection) unloadModule(module *Module) (err error) {
	if !module.loaded {
		return errors.New("Module not loaded: " + module.Name)
	}
	module.loaded = false
	if module.OnUnload != nil {
		module.OnUnload(conn)
	}
	return nil
}

//Loads a module
func (conn *Connection) Load(mod string) (err error) {
	module, exists := conn.modules[mod]
	if !exists {
		return errors.New("Unknown module: " + mod)
	}
	return conn.loadModule(module)
}

//Unloads a module
func (conn *Connection) Unload(mod string) (err error) {
	module, exists := conn.modules[mod]
	if !exists {
		return errors.New("Unknown module: " + mod)
	}
	return conn.unloadModule(module)
}

//Register a module for loading
func (conn *Connection) RegisterModule(module *Module) (err error) {
	_, exists := conn.modules[module.Name]
	if exists {
		return errors.New("Module already registered by that name: " + module.Name)
	}
	conn.modules[module.Name] = module
	return nil
}

//Registers all the modules in the list passed. Modules will only be registered
//if every module in the list is registered sucessfully
func (conn *Connection) RegisterModules(modules []*Module) (err error) {
	for i, module := range modules {
		if err := conn.RegisterModule(module); err != nil {
			//Unregister them because we had an error
			for _, mod := range modules[:i] {
				delete(conn.modules, mod.Name)
			}
			return err
		}
	}
	return nil
}
