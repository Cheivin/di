package van

type Van struct {
	defaults *store
	override *store
}

func New() *Van {
	separator := "."
	return &Van{defaults: newStore(separator), override: newStore(separator)}
}

func (v *Van) SetDefault(key string, value any) {
	v.defaults.Set(key, value)
}

func (v *Van) Set(key string, value any) {
	v.override.Set(key, value)
}

func (v *Van) Get(key string) (val any) {
	val = v.override.Get(key)
	if val == nil {
		val = v.defaults.Get(key)
	}
	return val
}

func (v *Van) GetAll() map[string]any {
	mergeMap := copyStringMap(v.override.GetAll())
	mergeStringMap(v.defaults.GetAll(), mergeMap)
	return mergeMap
}
