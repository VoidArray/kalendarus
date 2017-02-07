package backends

type Backend interface {
	Load(key string, vars interface{}) error
	Save(key string, vars interface{}) error
}
