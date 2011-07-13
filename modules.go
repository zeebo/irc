package irc

import "os"

type Module struct {
	Callbacks map[string]Callback
	Info      string
	Name      string
	loaded    bool
	OnLoad    func(*Connection)
	OnUnload  func(*Connection)
}

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

func (conn *Connection) Load(mod string) (err os.Error) {
	module, exists := conn.modules[mod]
	if !exists {
		return os.NewError("Unknown module: " + mod)
	}
	return conn.loadModule(module)
}

func (conn *Connection) Unload(mod string) (err os.Error) {
	module, exists := conn.modules[mod]
	if !exists {
		return os.NewError("Unknown module: " + mod)
	}
	return conn.unloadModule(module)
}

func (conn *Connection) RegisterModule(module *Module) (err os.Error) {
	_, exists := conn.modules[module.Name]
	if exists {
		return os.NewError("Module already registered by that name: " + module.Name)
	}
	conn.modules[module.Name] = module
	return nil
}
